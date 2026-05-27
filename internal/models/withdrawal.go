package models

import "time"

type Withdrawal struct {
	ID        int64
	UserId    int64
	OrderNum  int
	Amount    float32
	CreatedAt time.Time
}
