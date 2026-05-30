package models

import "time"

type Withdrawal struct {
	ID        int64
	UserID    int64
	OrderNum  int
	Amount    float32
	CreatedAt time.Time
}
