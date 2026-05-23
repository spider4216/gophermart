package handler

import "github.com/spider4216/gophermart/internal/models"

func (h Handler) mapOrdersResp(orders []models.Order) []models.OrderResp {
	var resp []models.OrderResp

	for _, order := range orders {
		resp = append(resp, models.OrderResp{
			Number:     order.Num,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UpdatedAt,
		})
	}

	return resp
}
