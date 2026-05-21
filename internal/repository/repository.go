package repository

import (
	"context"

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
