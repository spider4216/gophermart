package repository

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (r *Repository) GetTotalUserWithdrawn(ctx context.Context, userId int64) (float32, error) {
	return r.store.GetTotalUserWithdrawn(ctx, userId)
}

func (r *Repository) GetUserWithdrawals(ctx context.Context, userId int64) ([]models.Withdrawal, error) {
	return r.store.GetUserWithdrawals(ctx, userId)
}

func (r *Repository) Withdraw(ctx context.Context, userId int64, num int, amount float32) error {
	return r.store.Withdraw(ctx, userId, num, amount)
}
