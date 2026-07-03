package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	Subject    string `json:"sub"`
	TelegramID int64  `json:"telegram_id"`
	IsAdmin    bool   `json:"is_admin"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
}

func Sign(secret string, claims Claims) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("jwt secret is empty")
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerBytes) + "." + base64.RawURLEncoding.EncodeToString(claimsBytes)
	sig := sign(secret, unsigned)
	return unsigned + "." + sig, nil
}

func Verify(secret string, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	expected := sign(secret, unsigned)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

func NewClaims(userID int64, telegramID int64, isAdmin bool, ttl time.Duration) Claims {
	now := time.Now()
	return Claims{
		Subject:    strconv.FormatInt(userID, 10),
		TelegramID: telegramID,
		IsAdmin:    isAdmin,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(ttl).Unix(),
	}
}

func sign(secret, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func UserIDFromSubject(subject string) (int64, error) {
	id, err := strconv.ParseInt(subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid subject")
	}
	return id, nil
}
