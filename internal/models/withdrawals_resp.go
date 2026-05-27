package models

import "time"

type WithdrawalsResp struct {
	Order       int       `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
