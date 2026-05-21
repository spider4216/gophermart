package store

import (
	"context"

	"go.uber.org/zap"
)

const (
	PostgreDriver = "pgx"
)

type Storage interface {
	Ping(ctx context.Context) error
}

func New(driver string, dsn string, logger *zap.SugaredLogger) (Storage, error) {
	pgxStore, err := NewPgxStore(dsn, logger)
	if err != nil {
		return nil, err
	}

	return pgxStore, nil
}
