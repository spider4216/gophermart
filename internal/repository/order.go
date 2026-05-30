package repository

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (r *Repository) UpdateOrderStatus(ctx context.Context, orderNum int, userId int64, status models.OrderStatus, sum float32) error {
	return r.store.UpdateOrderStatus(ctx, orderNum, userId, status, sum)
}

func (r *Repository) GetOrdersByUserId(ctx context.Context, userId int64) ([]models.Order, error) {
	return r.store.GetUserOrders(ctx, userId)
}

func (r *Repository) GetOrderByUserId(ctx context.Context, num int, userId int64) (*models.Order, error) {
	return r.store.GetUserOrder(ctx, num, userId)
}

func (r *Repository) CreateOrder(ctx context.Context, userId int64, num int) (int64, error) {
	order := models.Order{
		UserID: userId,
		Num:    num,
		Status: models.OrderStatusNew,
	}

	return r.store.CreateOrder(ctx, order)
}
