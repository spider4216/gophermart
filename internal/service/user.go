package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/spider4216/gophermart/internal/models"
)

// Комплексный метод создания пользователя, включает след. операции
// - Создание нового пользователя
// - Создание нулевого баланса для нового пользователя
func (s *Service) CreateUser(ctx context.Context, username string, pass string) (int64, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, err
	}

	userId, err := s.SignUpUser(ctx, username, pass)
	if err != nil {
		if terr := tx.Rollback(); terr != nil {
			return 0, terr
		}

		return 0, err
	}

	if _, err = s.CreateUserBalance(ctx, userId); err != nil {
		if terr := tx.Rollback(); terr != nil {
			return 0, terr
		}

		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *Service) SignUpUser(ctx context.Context, username string, pass string) (int64, error) {
	hash := sha256.Sum256([]byte(pass))

	hashString := hex.EncodeToString(hash[:])

	id, err := s.repo.CreateUser(ctx, username, hashString)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.repo.GetUserByUsername(ctx, login)
}
