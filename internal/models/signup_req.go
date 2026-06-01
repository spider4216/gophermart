package models

type SignUpReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
