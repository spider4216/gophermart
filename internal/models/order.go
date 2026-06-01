package models

import "time"

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusProcess   OrderStatus = "PROCESSING"
	OrderStatusInvalid   OrderStatus = "INVALID"
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	ID        int64
	UserID    int64
	Num       int
	Status    OrderStatus
	Accrual   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
