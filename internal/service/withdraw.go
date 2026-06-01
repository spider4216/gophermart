package service

import (
	"context"

	"github.com/spider4216/gophermart/internal/models"
)

func (s *Service) GetUserBalanceWithWithdrawn(ctx context.Context, userId int64) (*models.Balance, float32, error) {
	// Получить текущий баланс пользователя
	balance, err := s.GetUserBalance(ctx, userId)
	if err != nil {
		return nil, 0, err
	}

	// Получить все списания бонусов пользователя
	withdrawn, err := s.repo.GetTotalUserWithdrawn(ctx, userId)
	if err != nil {
		return nil, 0, err
	}

	return balance, withdrawn, nil
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userId int64) ([]models.Withdrawal, error) {
	return s.repo.GetUserWithdrawals(ctx, userId)
}

// Метод списания бонусов
// - Получает текущий баланс пользователя
// - Проверяет возможность списания через остаток
// - Создает списание
// - Обновляет баланс
func (s *Service) Withdraw(ctx context.Context, userId int64, num int, amount float32) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}

	balance, err := s.GetUserBalance(ctx, userId)
	if err != nil {
		if terr := tx.Rollback(); terr != nil {
			return terr
		}

		return err
	}

	remain := balance.Balance - amount

	s.logger.Debug("withdraw: ", amount, " user ", userId, " balance: ", balance.Balance, " remain: ", remain)

	// Если баланса недостаточно
	if remain < 0 {
		// Свой тип ошибки
		s.logger.Error("No balance for withdraw ", amount, " Current balance ", balance.Balance)
		return &NoBalance{}
	}

	// Создаем списание
	if err := s.repo.Withdraw(ctx, userId, num, amount); err != nil {
		if terr := tx.Rollback(); terr != nil {
			return terr
		}

		return err
	}

	// Обновляем баланс
	if err := s.repo.UpdateUserBalance(ctx, userId, remain); err != nil {
		if terr := tx.Rollback(); terr != nil {
			return terr
		}

		return err
	}

	return tx.Commit()
}
