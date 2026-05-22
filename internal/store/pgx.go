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

func (db *PgxStore) CreateUser(ctx context.Context, user models.User) (int64, error) {
	sql := "INSERT INTO users (username,password) VALUES ($1,$2)"

	res, err := db.DB.ExecContext(ctx, sql, &user.Username, &user.Password)
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
	sql := "SELECT username,password FROM users WHERE username=$1"

	user := models.User{}

	err := db.DB.QueryRow(sql, username).Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
