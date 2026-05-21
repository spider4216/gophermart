package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/models"
	"github.com/spider4216/gophermart/internal/service"
	"go.uber.org/zap"
)

func New(cfg *config.Config, logger *zap.SugaredLogger, service service.Service) Handler {
	return Handler{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}
}

type Handler struct {
	cfg     *config.Config
	service service.Service
	logger  *zap.SugaredLogger
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
	id, err := h.service.SignUpUser(ctx, req.Login, req.Password)
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

	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}
