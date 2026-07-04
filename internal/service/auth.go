package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"sakeofher/internal/domain"
	appjwt "sakeofher/internal/platform/jwt"
)

type authService struct {
	adminUsername string
	adminPassword string
	jwtSecret     string
	accessTTL     time.Duration
}

func NewAuthService(
	adminUsername string,
	adminPassword string,
	jwtSecret string,
	accessTTL time.Duration,
) AuthService {
	return &authService{
		adminUsername: adminUsername,
		adminPassword: adminPassword,
		jwtSecret:     jwtSecret,
		accessTTL:     accessTTL,
	}
}

func (s *authService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthSession, error) {
	_ = ctx

	username := strings.TrimSpace(input.Login)
	if username == "" {
		username = strings.TrimSpace(input.Username)
	}

	password := input.Password

	if username == "" || password == "" {
		return nil, domain.ErrInvalidInput
	}

	if strings.TrimSpace(s.adminUsername) == "" || s.adminPassword == "" {
		return nil, domain.ErrUnauthorized
	}

	if !secureCompare(username, s.adminUsername) || !secureCompare(password, s.adminPassword) {
		return nil, domain.ErrUnauthorized
	}

	jwtClaims := appjwt.NewAdminClaims(username, s.accessTTL)

	accessToken, err := appjwt.Sign(s.jwtSecret, jwtClaims)
	if err != nil {
		return nil, err
	}

	return &domain.AuthSession{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   time.Unix(jwtClaims.ExpiresAt, 0),
		Username:    username,
		IsAdmin:     true,
	}, nil
}

func secureCompare(given, expected string) bool {
	givenHash := sha256.Sum256([]byte(given))
	expectedHash := sha256.Sum256([]byte(expected))

	return subtle.ConstantTimeCompare(givenHash[:], expectedHash[:]) == 1
}
