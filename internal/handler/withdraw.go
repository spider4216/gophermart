package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/spider4216/gophermart/internal/models"
	"github.com/spider4216/gophermart/internal/service"
	"go.uber.org/zap"
)

// Списание бонусов
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

	var ErrNoBalance *service.NoBalance

	err = h.service.Withdraw(ctx, userId, orderNum, req.Sum)

	if errors.As(err, &ErrNoBalance) {
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

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		h.logger.Error("Cannot write response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
