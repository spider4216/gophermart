package service

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (s *Service) IncreaseUserBalance(ctx context.Context, userId int64, amount float32) error {
	b, err := s.GetUserBalance(ctx, userId)
	if err != nil {
		return err
	}

	sum := b.Balance + amount

	return s.UpdateUserBalance(ctx, userId, sum)
}

func (s *Service) UpdateUserBalance(ctx context.Context, userId int64, amount float32) error {
	return s.repo.UpdateUserBalance(ctx, userId, amount)
}

func (s *Service) GetUserBalance(ctx context.Context, userId int64) (*models.Balance, error) {
	return s.repo.GetUserBalance(ctx, userId)
}

func (s *Service) CreateUserBalance(ctx context.Context, userId int64) (int64, error) {
	s.logger.Debug("Create balance for user ", userId)
	return s.repo.CreateUserBalance(ctx, userId)
}
