package model

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	Balance      float64   `json:"balance" db:"balance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type UserLoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
