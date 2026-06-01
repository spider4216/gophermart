package handler

import (
	"strconv"

	"github.com/spider4216/gophermart/internal/models"
)

func (h Handler) mapOrdersResp(orders []models.Order) []models.OrderResp {
	return mapResp(orders, func(order models.Order) models.OrderResp {
		return models.OrderResp{
			Number:     strconv.Itoa(order.Num),
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UpdatedAt,
		}
	})
}

func (h Handler) mapWithdrawalsResp(withdrawals []models.Withdrawal) []models.WithdrawalsResp {
	return mapResp(withdrawals, func(withdrawal models.Withdrawal) models.WithdrawalsResp {
		return models.WithdrawalsResp{
			Order:       strconv.Itoa(withdrawal.OrderNum),
			Sum:         withdrawal.Amount,
			ProcessedAt: withdrawal.CreatedAt,
		}
	})
}

func mapResp[T any, R any](items []T, fn func(T) R) []R {
	var res []R

	for _, item := range items {
		res = append(res, fn(item))
	}

	return res
}
