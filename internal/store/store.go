package store

import (
	"context"
	"database/sql"

	"github.com/spider4216/gophermart/internal/models"
	"go.uber.org/zap"
)

const (
	PostgreDriver = "pgx"
)

type Storage interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user models.User) (int64, error)
	GetUser(ctx context.Context, username string) (*models.User, error)
	CreateOrder(ctx context.Context, order models.Order) (int64, error)
	GetUserOrder(ctx context.Context, num int, userId int64) (*models.Order, error)
	GetUserOrders(ctx context.Context, userId int64) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNum int, userId int64, status models.OrderStatus, sum float32) error
	CreateUserBalance(ctx context.Context, userId int64) (int64, error)
	UpdateUserBalance(ctx context.Context, userId int64, amount float32) error
	GetUserBalance(ctx context.Context, userId int64) (*models.Balance, error)
	Withdraw(ctx context.Context, userId int64, num int, amount float32) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	GetUserWithdrawals(ctx context.Context, userId int64) ([]models.Withdrawal, error)
	GetTotalUserWithdrawn(ctx context.Context, userId int64) (float32, error)
}

func New(driver string, dsn string, logger *zap.SugaredLogger) (Storage, error) {
	pgxStore, err := NewPgxStore(dsn, logger)
	if err != nil {
		return nil, err
	}

	return pgxStore, nil
}
