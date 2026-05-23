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
	sql := "INSERT INTO orders (user_id,num,status) VALUES ($1,$2,$3)"

	res, err := db.DB.ExecContext(ctx, sql, order.UserId, order.Num, order.Status)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (db *PgxStore) CreateUser(ctx context.Context, user models.User) (int64, error) {
	sql := "INSERT INTO users (username,password) VALUES ($1,$2)"

	res, err := db.DB.ExecContext(ctx, sql, user.Username, user.Password)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, nil
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
