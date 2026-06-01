package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// Получение списка заказов пользователя
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

	w.Header().Add("Content-Type", "application/json")
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
