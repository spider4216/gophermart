package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/models"
	"github.com/spider4216/gophermart/internal/service"
	"go.uber.org/zap"
)

func New(cfg *config.Config, logger *zap.SugaredLogger, service *service.Service) Handler {
	return Handler{
		cfg:       cfg,
		service:   service,
		logger:    logger,
		semaphore: make(chan struct{}, cfg.RegMaxPool),
	}
}

type Handler struct {
	cfg       *config.Config
	service   *service.Service
	logger    *zap.SugaredLogger
	semaphore chan struct{}
}

func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot ping store", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Ping store OK")
}

func (h Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	userId := h.service.GetUserIdFromCtx(ctx)

	h.logger.Debug("User ID ", userId)

	// Получаем тело запроса
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := models.WithdrawReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed read unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNum, err := strconv.Atoi(req.Order)
	if err != nil {
		h.logger.Error("Order is not number", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Валидирую номер заказа
	if !h.service.IsOrderNumValid(orderNum) {
		h.logger.Error("Invalid order number", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var ErrNoBalance service.NoBalance

	err = h.service.Withdraw(ctx, userId, orderNum, req.Sum)

	if errors.Is(err, &ErrNoBalance) {
		h.logger.Error("No balance", zap.Error(err))
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	if err != nil {
		h.logger.Error("Something went wrong", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// История списания бонусов у пользователя
func (h Handler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	userId := h.service.GetUserIdFromCtx(ctx)

	h.logger.Debug("User ID ", userId)

	withdrawals, err := h.service.GetUserWithdrawals(ctx, userId)
	if err != nil {
		h.logger.Error("cannot get user withdrawals", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) <= 0 {
		h.logger.Error("No withdrawals found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := h.mapWithdrawalsResp(withdrawals)

	b, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("cannot marshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		h.logger.Error("Cannot write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	userId := h.service.GetUserIdFromCtx(ctx)

	h.logger.Debug("User ID ", userId)

	orders, err := h.service.GetOrdersByUserId(ctx, userId)
	if err != nil {
		h.logger.Error("cannot get user orders", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) <= 0 {
		h.logger.Error("No orders found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := h.mapOrdersResp(orders)

	b, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("cannot marshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		h.logger.Error("Cannot write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Загрузка номера заказа
func (h Handler) RegOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	h.logger.Debug("User ID ", h.service.GetUserIdFromCtx(ctx))

	// Получаем тело запроса
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразую номер заказа в число
	num, err := strconv.Atoi(string(body))
	if err != nil {
		h.logger.Error("cannot convert order number to int", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Валидирую номер заказа
	if !h.service.IsOrderNumValid(num) {
		h.logger.Error("Invalid order number", zap.Error(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userId := h.service.GetUserIdFromCtx(ctx)

	order, err := h.service.GetOrderByUserId(ctx, num, userId)
	if err != nil {
		// Если ошибка не связана с отсутствием записей в таблице
		if !errors.Is(err, sql.ErrNoRows) {
			h.logger.Error("cannot get order", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if order != nil {
		// Если номер заказа уже был ранее зарегистрирован у пользователя
		// то прекращаем работу
		h.logger.Error("user has already have order number: ", num)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Создаю заказ
	orderId, err := h.service.CreateOrder(ctx, userId, num)
	if err != nil {
		// Если ошибка является дубликатом
		if err != nil && h.service.IsErrAsDuplicate(err) {
			h.logger.Error("Order has already exist", zap.Error(err))
			w.WriteHeader(http.StatusConflict)
			return
		}

		h.logger.Error("cannot create order", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Order was created: ", orderId)

	withoutCancel := context.WithoutCancel(r.Context())

	select {
	case h.semaphore <- struct{}{}:
		go func() {
			defer func() {
				h.logger.Debug("Release pool in reg order")
				<-h.semaphore
			}()
			if err := h.service.CalcBonus(withoutCancel, num); err != nil {
				h.logger.Error("Cannot calc", zap.Error(err))
			}
		}()
	default:
		h.logger.Error("Too many requests for reg orders.")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Авторизация пользователя
func (h Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	// Получаем тело запроса
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := models.SignInReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed read unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлечение пользователя по логину
	user, err := h.service.GetUserByLogin(ctx, req.Login)
	if err != nil {
		h.logger.Error("User not found", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if !h.service.CheckPass(user, req.Pass) {
		h.logger.Error("password is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Авторизация
	token, err := h.service.BuildJWTString(user.ID, h.cfg.SecretKey, h.cfg.ExpToken)
	if err != nil {
		h.logger.Error("Unathorized", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Add("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

// Регистрация пользователя и мгновенная его авторизация
func (h Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	// Получаем тело запроса
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := models.SugnUpReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed read unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Регистрация пользователя
	id, err := h.service.CreateUser(ctx, req.Login, req.Password)
	if err != nil {
		// Если ошибка является дубликатом
		if err != nil && h.service.IsErrAsDuplicate(err) {
			h.logger.Error("User have already exist", zap.Error(err))
			w.WriteHeader(http.StatusConflict)
			return
		}

		h.logger.Error("cannot create user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Авторизация
	token, err := h.service.BuildJWTString(id, h.cfg.SecretKey, h.cfg.ExpToken)
	if err != nil {
		h.logger.Error("Unathorized", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Add("Authorization", token)
	w.WriteHeader(http.StatusOK)
}
