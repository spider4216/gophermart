package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spider4216/gophermart/internal/models"
	"go.uber.org/zap"
)

// Получение баланса пользователя с его списаниями
func (h Handler) GetUserBalanceAndWithdrawn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.CtxTimeout)

	defer cancel()

	userId := h.service.GetUserIdFromCtx(ctx)

	h.logger.Debug("User ID ", userId)

	balance, withdrawn, err := h.service.GetUserBalanceWithWithdrawn(ctx, userId)
	if err != nil {
		h.logger.Error("Something went wrong", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := models.BalanceResp{
		Current:   balance.Balance,
		Withdrawn: withdrawn,
	}

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
