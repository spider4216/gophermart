package store

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
	"go.uber.org/zap"
)

const (
	PostgreDriver = "pgx"
)

type Storage interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user models.User) (int64, error)
}

func New(driver string, dsn string, logger *zap.SugaredLogger) (Storage, error) {
	pgxStore, err := NewPgxStore(dsn, logger)
	if err != nil {
		return nil, err
	}

	return pgxStore, nil
}
