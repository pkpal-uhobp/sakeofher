package domain

import "time"

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthSession struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	Username    string    `json:"username"`
	IsAdmin     bool      `json:"is_admin"`
}
