package models

type SignInReq struct {
	Login string `json:"login"`
	Pass  string `json:"password"`
}
