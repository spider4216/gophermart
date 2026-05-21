package handler

import (
	"context"
	"net/http"

	"github.com/spider4216/gophermart/internal/config"
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
	ctx, cancel := context.WithTimeout(context.Background(), h.cfg.CtxTimeout)

	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot ping store", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Ping store OK")
}
