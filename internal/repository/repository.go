package repository

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
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

func (r *Repository) CreateUser(ctx context.Context, username string, hash string) (int64, error) {
	user := models.User{
		Username: username,
		Password: hash,
	}

	return r.store.CreateUser(ctx, user)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.store.GetUser(ctx, username)
}
