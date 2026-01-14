package model

type UserSession struct {
	ID       int64  `json:"id,string"`
	Username string `json:"username"`
}
