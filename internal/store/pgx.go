package store

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgxStore struct {
	DB     *sql.DB
	logger *zap.SugaredLogger
}

func NewPgxStore(dsn string, logger *zap.SugaredLogger) (*PgxStore, error) {
	db, err := sql.Open(PostgreDriver, dsn)

	if err != nil {
		return nil, err
	}

	return &PgxStore{DB: db, logger: logger}, nil
}

func (db *PgxStore) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
