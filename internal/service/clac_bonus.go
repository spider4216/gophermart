package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	StatusRegistered string = "REGISTERED"
	StatusProcessing string = "PROCESSING"
	StatusInvalid    string = "INVALID"
	StatusProcessed  string = "PROCESSED"
)

// Обертка ответа от внешнего сервиса
type RemoteResp struct {
	StatusCode int
	Retry      time.Duration
	Data       RemoteData
}

// Тело ответа от внешнего сервиса
type RemoteData struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (s *Service) CalcBonus(ctx context.Context, num int) error {
	// 1. Отправка запроса в систему лояльности
	// GET /api/orders/{number}
	// 2. Ожидание финального статуса
	// Финальные:
	// - INVALID - обновить статус в таблице и прекратить опрос
	// - PROCESSED - обновить статус и начислить бонус и прекратить опрос
	//
	// Не финальные статусы
	// - REGISTERED - ничего не обновлять, продолжать опрос
	// - PROCESSING - Обновить со статуса NEW -> PROCESSING и продолжать опрос
	// Негативные сценарии
	// - HTTP 204 заказ не зарегистрирован в системе расчёта
	// - HTTP 429 превышено количество запросов к сервису.
	// При HTTP 429 Too Many Requests
	// - Обратить внимание на заголовок Retry-After
	// - Прервать опрос со всех гоурутин
	// - Возобновить опрос во всех гоурутинах через Retry-After

	userId := s.GetUserIdFromCtx(ctx)

	s.logger.Debug("Start calcing for ", num, " user id ", userId)

	// Создаем таймер
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		// Перед каждым тиком проверяем нет ли глобальной паузы
		// Которая запускается при Retry Timeout
		// Если пауза установлена, то ожидаем разблокировки
		if err := s.checkPause(ctx); err != nil {
			return err
		}

		select {
		case <-ticker.C:
			// Таймер сработал, отправляем запрос
			s.logger.Debug("Send req to remote")
			resp, err := s.sendReq(num)
			if err != nil {
				return err
			}

			// Если уперлись в лимит времени, то ожидаем
			if resp.Retry != 0 {
				s.logger.Debug("Too many requests to remote, wait ", resp.Retry, "s")
				// Тут устаналиваем глобальную паузу, все горутины начнут
				// ожидать пока пауза не истечет. В методе startGlobalPause в отдельноу
				// горутине запустится таймер по истечении которого канал будет оповещен
				// и цикл перестанет блокироваться в самом начале
				s.startGlobalPause(resp.Retry)
				continue
			}

			// Если 204, то заказ не зарегистрирован
			if resp.StatusCode == http.StatusNoContent {
				s.logger.Error("Order was not registered")
				return fmt.Errorf("order was not registered")
			}

			// Если статус не 200, то какая-то ошибка
			if resp.StatusCode != http.StatusOK {
				s.logger.Error("Something went wrong on remote side")
				return fmt.Errorf("something went wrong on remote side")
			}

			// Заказ обработан успешно, начинаем отслеживать статусы
			// Если статус REGISTERED, пропускаем итерацию
			if resp.Data.Status == StatusRegistered {
				s.logger.Debug("Status REGISTERED, try again...")
				continue
			}

			// Если статус PROCESSING, то обновляем статус в БД и продолжаем опрос
			if resp.Data.Status == StatusProcessing {
				s.logger.Debug("Status PROCESSING, try again...")
				if err := s.UpdateOrderProcess(ctx, num, userId); err != nil {
					return err
				}

				continue
			}

			// Если статус финальный но ошибочный - INVALID, то обновляем статус в БД
			// и прекращаем работу
			if resp.Data.Status == StatusInvalid {
				s.logger.Debug("Status INVALID, exit...")
				if err := s.UpdateOrderInvalid(ctx, num, userId); err != nil {
					return err
				}

				return nil
			}

			// Если сьатус финальный и не ошибочный - PROCESSED, то обновляем в БД
			// и прекращаем работу
			if resp.Data.Status == StatusProcessed {
				s.logger.Debug("Status PROCESSED, done, exit...")
				if err := s.UpdateOrderProcessed(ctx, num, userId, resp.Data.Accrual); err != nil {
					return err
				}

				return nil
			}

		case <-ctx.Done():
			s.logger.Debug("Context in remote polling was canceled")
			return fmt.Errorf("context wac canceled")
		}
	}
}

func (s *Service) sendReq(num int) (*RemoteResp, error) {
	numStr := strconv.Itoa(num)
	// Готовим URL
	fullUrl, err := url.JoinPath(s.cfg.AccrualAddr, "/api/orders/", numStr)
	if err != nil {
		return nil, err
	}

	remote := RemoteData{}

	client := s.httpC

	s.logger.Debug("Send to ", fullUrl)

	// Отправляем запрос на внешний сервис и мапим ответ
	resp, err := client.R().
		SetResult(&remote).
		Get(fullUrl)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("Remote response ", string(resp.Body()))

	// Готовим обертку с ответом
	res := RemoteResp{
		StatusCode: resp.StatusCode(),
		Data:       remote,
	}

	// Если попали в временной лимит запросов
	if resp.StatusCode() == http.StatusTooManyRequests {
		// Вносим в обертку данные о том, сколько нужно подождать
		retryStr := resp.Header().Get("Retry-After")
		retry, err := strconv.Atoi(retryStr)
		if err != nil {
			return nil, err
		}

		res.Retry = time.Duration(retry) * time.Second
	}

	return &res, nil
}

// Проверяет активна ли пауза и если да, горутина будет ждать ее окончания
func (s *Service) checkPause(ctx context.Context) error {
	s.pauseMu.RLock()

	// Если глобальная пауза ранее никем не была включена
	// То выходим без ожидания
	if !s.isPaused {
		s.pauseMu.RUnlock()
		return nil
	}

	// Если пауза активна, то ожидаем
	// пока пауза не прекратится через канал или
	// пока контекст не отменится
	ch := s.pauseChan
	s.pauseMu.RUnlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) startGlobalPause(duration time.Duration) {
	s.pauseMu.Lock()

	// Если кто-то включил глобальную паузу ранее, то
	// нет необходимости ее снова включать
	if s.isPaused {
		s.pauseMu.Unlock()
		return
	}

	s.logger.Debug("Start global pause on ", duration)

	s.isPaused = true
	// Создаем глобальный канал который будут все ждать
	s.pauseChan = make(chan struct{})
	s.pauseMu.Unlock()

	// Чтобы не блокировать текущий метод, ожидание запускаем
	// в Go рутине, она сама запишет в канал когда
	// таймаут пройдет. Примерно так мы уже делали в одном из инкрементов

	go func() {
		time.Sleep(duration)

		s.pauseMu.Lock()
		// Снимаем глобальную паузу
		s.isPaused = false
		// Закрываем глобальный канал, тут поидее все горутины
		// которые его слушали должны разблокироваться
		close(s.pauseChan)
		s.pauseMu.Unlock()

		s.logger.Debug("Global pause ended")
	}()
}
