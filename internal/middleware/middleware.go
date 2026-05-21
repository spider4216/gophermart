package middleware

import (
	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/service"
	"go.uber.org/zap"
)

type Middleware struct {
	logger  *zap.SugaredLogger
	cfg     *config.Config
	service service.Service
}

func New(logger *zap.SugaredLogger, cfg *config.Config, service service.Service) Middleware {
	return Middleware{
		logger:  logger,
		cfg:     cfg,
		service: service,
	}
}
