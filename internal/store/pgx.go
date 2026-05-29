package store

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spider4216/gophermart/internal/models"
)

type PgxStore struct {
	DB     *sql.DB
	logger *zap.SugaredLogger
}

func NewPgxStore(dsn string, logger *zap.SugaredLogger) (*PgxStore, error) {
	db, err := sql.Open(PostgreDriver, dsn)
	if err != nil {
		return nil, err
	}

	return &PgxStore{DB: db, logger: logger}, nil
}

func (db *PgxStore) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db *PgxStore) GetUserBalance(ctx context.Context, userId int64) (*models.Balance, error) {
	sql := "SELECT id,user_id,amount FROM balances WHERE user_id=$1"

	balance := models.Balance{}

	err := db.DB.QueryRow(sql, userId).Scan(&balance.ID, &balance.UserID, &balance.Balance)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (db *PgxStore) UpdateUserBalance(ctx context.Context, userId int64, amount float32) error {
	sql := "UPDATE balances SET amount=$1 WHERE user_id=$2"

	if _, err := db.DB.ExecContext(ctx, sql, amount, userId); err != nil {
		return err
	}

	return nil
}

func (db *PgxStore) CreateUserBalance(ctx context.Context, userId int64) (int64, error) {
	sql := "INSERT INTO balances (user_id,amount) VALUES ($1,0) RETURNING id"

	var id int64

	if err := db.DB.QueryRowContext(ctx, sql, userId).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *PgxStore) UpdateOrderStatus(ctx context.Context, orderNum int, userId int64, status models.OrderStatus, sum float32) error {
	sql := "UPDATE orders SET status=$1,accrual=$2 WHERE user_id=$3 AND num=$4"

	if _, err := db.DB.ExecContext(ctx, sql, status, sum, userId, orderNum); err != nil {
		return err
	}

	return nil
}

func (db *PgxStore) GetUserOrders(ctx context.Context, userId int64) ([]models.Order, error) {
	sql := "SELECT id,user_id,num,status,accrual,created_at,updated_at FROM orders WHERE user_id=$1"

	rows, err := db.DB.QueryContext(ctx, sql, userId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			db.logger.Warn("Cannot close rows", zap.Error(err))
		}
	}()

	var items []models.Order

	for rows.Next() {
		var item models.Order

		if err := rows.Scan(
			&item.ID,
			&item.UserId,
			&item.Num,
			&item.Status,
			&item.Accrual,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (db *PgxStore) GetUserOrder(ctx context.Context, num int, userId int64) (*models.Order, error) {
	sql := "SELECT id,user_id,num,status,accrual,created_at,updated_at FROM orders WHERE num=$1 AND user_id=$2"

	order := models.Order{}

	err := db.DB.QueryRow(sql, num, userId).Scan(
		&order.ID,
		&order.UserId,
		&order.Num,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (db *PgxStore) CreateOrder(ctx context.Context, order models.Order) (int64, error) {
	sql := "INSERT INTO orders (user_id,num,status) VALUES ($1,$2,$3) RETURNING id"

	var id int64

	if err := db.DB.QueryRowContext(ctx, sql, order.UserId, order.Num, order.Status).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *PgxStore) CreateUser(ctx context.Context, user models.User) (int64, error) {
	sql := "INSERT INTO users (username,password) VALUES ($1,$2) RETURNING id"

	var id int64

	if err := db.DB.QueryRowContext(ctx, sql, user.Username, user.Password).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (db *PgxStore) GetUser(ctx context.Context, username string) (*models.User, error) {
	sql := "SELECT id,username,password FROM users WHERE username=$1"

	user := models.User{}

	err := db.DB.QueryRow(sql, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *PgxStore) Withdraw(ctx context.Context, userId int64, num int, amount float32) error {
	sql := "INSERT INTO withdrawals (user_id,order_num,amount) VALUES ($1,$2,$3)"

	if _, err := db.DB.ExecContext(ctx, sql, userId, num, amount); err != nil {
		return err
	}

	return nil
}

func (db *PgxStore) GetUserWithdrawals(ctx context.Context, userId int64) ([]models.Withdrawal, error) {
	sql := "SELECT id,user_id,order_num,amount,created_at FROM withdrawals WHERE user_id=$1 ORDER BY created_at DESC"

	rows, err := db.DB.QueryContext(ctx, sql, userId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			db.logger.Warn("Cannot close rows", zap.Error(err))
		}
	}()

	var items []models.Withdrawal

	for rows.Next() {
		var item models.Withdrawal

		if err := rows.Scan(
			&item.ID,
			&item.UserId,
			&item.OrderNum,
			&item.Amount,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (db *PgxStore) GetTotalUserWithdrawn(ctx context.Context, userId int64) (float32, error) {
	sql := "SELECT COALESCE(SUM(amount), 0) FROM withdrawals WHERE user_id=$1"

	var sum float32

	err := db.DB.QueryRow(sql, userId).Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, nil
}

func (db *PgxStore) BeginTx(ctx context.Context) (Tx, error) {
	return db.DB.BeginTx(ctx, nil)
}

func (db *PgxStore) GetDB() *sql.DB {
	return db.DB
}
