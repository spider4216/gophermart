package repository

import (
	"context"
	"database/sql"

	"github.com/spider4216/gophermart/internal/store"
)

func New(store store.Storage) *Repository {
	return &Repository{
		store: store,
	}
}

type Repository struct {
	store store.Storage
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.store.Ping(ctx)
}

func (r *Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.store.BeginTx(ctx)
}
