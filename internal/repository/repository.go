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
		UserId: userId,
		Num:    num,
		Status: models.OrderStatusNew,
	}

	return r.store.CreateOrder(ctx, order)
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
