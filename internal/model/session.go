package model

import "time"

const (
	SessionCookieToken = "Token"
	SessionExpireDays  = 7
)

type Session struct {
	Token          string    `json:"token" db:"token"`
	UserID         int       `json:"user_id" db:"user_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	ExpireAt       time.Time `json:"expire_at" db:"expire_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
}
