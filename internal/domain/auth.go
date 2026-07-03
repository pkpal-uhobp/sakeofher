package domain

import "time"

type TelegramOAuthStart struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

type TelegramOAuthCallbackInput struct {
	Code          string `json:"code"`
	State         string `json:"state"`
	ExpectedState string `json:"-"`
	CodeVerifier  string `json:"-"`
	Nonce         string `json:"-"`
}

type TelegramOIDCTokenRequest struct {
	Code         string
	RedirectURI  string
	CodeVerifier string
}

type TelegramOIDCTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

type TelegramOIDCClaims struct {
	Issuer            string `json:"iss"`
	Audience          any    `json:"aud"`
	Subject           string `json:"sub"`
	IssuedAt          int64  `json:"iat"`
	ExpiresAt         int64  `json:"exp"`
	Nonce             string `json:"nonce,omitempty"`
	TelegramID        int64  `json:"id"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Picture           string `json:"picture,omitempty"`
	PhoneNumber       string `json:"phone_number,omitempty"`
}

type AuthSession struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        *User     `json:"user"`
	IsAdmin     bool      `json:"is_admin"`
}
