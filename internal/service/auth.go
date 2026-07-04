package service

import (
	"context"
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
	username := strings.TrimSpace(input.Username)
	password := input.Password

	if username == "" || password == "" {
		return nil, domain.ErrInvalidInput
	}
	if strings.TrimSpace(s.adminUsername) == "" || s.adminPassword == "" {
		return nil, domain.ErrUnauthorized
	}
	if username != s.adminUsername || !secureCompare(password, s.adminPassword) {
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
	return subtle.ConstantTimeCompare([]byte(given), []byte(expected)) == 1
}
