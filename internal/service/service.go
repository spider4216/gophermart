package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spider4216/gophermart/internal/repository"
	"go.uber.org/zap"
)

func New(repo *repository.Repository, logger *zap.SugaredLogger) Service {
	return Service{
		repo:   repo,
		logger: logger,
	}
}

type Service struct {
	repo   *repository.Repository
	logger *zap.SugaredLogger
}

func (s Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s Service) SignUpUser(ctx context.Context, username string, pass string) (int64, error) {
	hash := sha256.Sum256([]byte(pass))

	hashString := hex.EncodeToString(hash[:])

	id, err := s.repo.CreateUser(ctx, username, hashString)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s Service) IsErrAsDuplicate(err error) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.UniqueViolation
}
