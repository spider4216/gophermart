package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
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

func (s Service) CalcBonus(ctx context.Context, num int) (float32, error) {
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

	s.logger.Debug("Start calcing for ", num)

	// Создаем таймер
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Таймер сработал, отправляем запрос
			s.logger.Debug("Send req to remote")
			resp, err := s.sendReq(num)

			if err != nil {
				return 0, err
			}

			// Если уперлись в лимит времени, то ожидаем
			if resp.Retry != 0 {
				s.logger.Debug("Too many requests to remote, wait ", resp.Retry, "s")
				time.Sleep(resp.Retry * time.Second)
				// Затем перезапускаем итерацию
				// todo но тут еще должно бть что-то заставляет ожидать через канал все гоурутины
				// поскольку внешний сервис по условию задачи глобально встал для
				// всех заказов
				continue
			}

			// Если 204, то заказ не зарегистрирован
			if resp.StatusCode == http.StatusNoContent {
				s.logger.Error("Order was not registered")
				return 0, fmt.Errorf("Order was not registered")
			}

			// Если статус не 200, то какая-то ошибка
			if resp.StatusCode != http.StatusOK {
				s.logger.Error("Something went wrong on remote side")
				return 0, fmt.Errorf("Something went wrong on remote side")
			}

			// Заказ обработан успешно, начинаем отслеживать статусы
			// Если статус REGISTERED, пропускаем итерацию
			if resp.Data.Status == StatusRegistered {
				s.logger.Debug("Status REGISTERED, try again...")
				continue
			}

			// Если статус PROCESSING, то обновляем статус в БД и продолжаем опрос
			if resp.Data.Status == StatusProcessing {
				// todo update db status
				s.logger.Debug("Status PROCESSING, try again...")
				continue
			}

			// Если статус финальный но ошибочный - INVALID, то обновляем статус в БД
			// и прекращаем работу
			if resp.Data.Status == StatusInvalid {
				// todo update db status
				// Ошибки нет, если вернулась структура но сумма нулевая, значит INVALID
				s.logger.Debug("Status INVALID, exit...")
				return 0, nil
			}

			// Если сьатус финальный и не ошибочный - PROCESSED, то обновляем в БД
			// и прекращаем работу
			if resp.Data.Status == StatusProcessed {
				// todo update db status
				s.logger.Debug("Status PROCESSED, done, exit...")
				return resp.Data.Accrual, nil
			}

		case <-ctx.Done():
			s.logger.Debug("Context in remote polling was canceled")
			return 0, fmt.Errorf("Context wac canceled")
		}
	}
}

func (s Service) sendReq(num int) (*RemoteResp, error) {
	numStr := strconv.Itoa(num)
	// Готовим URL
	fullUrl, err := url.JoinPath(s.cfg.AccrualAddr, "/api/orders/", numStr)

	if err != nil {
		return nil, err
	}

	remote := RemoteData{}

	client := resty.New()

	s.logger.Debug("Send to ", fullUrl)

	// Отправляем запрос на внешний сервис и мапим ответ
	resp, err := client.R().
		SetResult(&remote).
		Get(fullUrl)

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

		res.Retry = time.Duration(retry)
	}

	return &res, nil
}
