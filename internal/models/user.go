package models

import "time"

type User struct {
	ID        int64
	Username  string
	Password  string
	CreatedAt time.Time
}
