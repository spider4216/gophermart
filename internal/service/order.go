package service

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (s *Service) UpdateOrderInvalid(ctx context.Context, orderNum int, userId int64) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusInvalid, 0)
}

func (s *Service) UpdateOrderProcessed(ctx context.Context, orderNum int, userId int64, sum float32) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusProcessed, sum)
}

func (s *Service) UpdateOrderProcess(ctx context.Context, orderNum int, userId int64) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusProcess, 0)
}

func (s *Service) GetOrdersByUserId(ctx context.Context, userId int64) ([]models.Order, error) {
	return s.repo.GetOrdersByUserId(ctx, userId)
}

func (s *Service) GetOrderByUserId(ctx context.Context, num int, userId int64) (*models.Order, error) {
	return s.repo.GetOrderByUserId(ctx, num, userId)
}

func (s *Service) CreateOrder(ctx context.Context, userId int64, num int) (int64, error) {
	return s.repo.CreateOrder(ctx, userId, num)
}
