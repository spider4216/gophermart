package models

type WithdrawReq struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
