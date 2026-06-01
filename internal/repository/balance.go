package repository

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (r *Repository) GetUserBalance(ctx context.Context, userId int64) (*models.Balance, error) {
	return r.store.GetUserBalance(ctx, userId)
}

func (r *Repository) UpdateUserBalance(ctx context.Context, userId int64, amount float32) error {
	return r.store.UpdateUserBalance(ctx, userId, amount)
}

func (r *Repository) CreateUserBalance(ctx context.Context, userId int64) (int64, error) {
	return r.store.CreateUserBalance(ctx, userId)
}
