package service

import (
	"context"

	"github.com/spider4216/gophermart/internal/repository"
	"go.uber.org/zap"
)

func New(repo *repository.Repository, logger *zap.SugaredLogger) Service {
	return Service{
		repo:   repo,
		logger: logger,
	}
}

type Service struct {
	repo   *repository.Repository
	logger *zap.SugaredLogger
}

func (s Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
