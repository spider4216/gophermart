package models

import "time"

const (
	New       OrderStatus = "NEW"
	Process   OrderStatus = "PROCESSING"
	Invalid   OrderStatus = "INVALID"
	Processed OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	ID        int64
	UserId    int64
	Num       int
	Status    OrderStatus
	Accrual   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
