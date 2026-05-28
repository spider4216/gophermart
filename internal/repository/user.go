package repository

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

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
