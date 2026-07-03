package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway"
	appjwt "sakeofher/internal/platform/jwt"
	"sakeofher/internal/repository"
)

type authService struct {
	repo         *repository.Repositories
	telegramOIDC gateway.TelegramOAuthGateway
	jwtSecret    string
	accessTTL    time.Duration
	redirectURI  string
}

func NewAuthService(
	repo *repository.Repositories,
	telegramOIDC gateway.TelegramOAuthGateway,
	jwtSecret string,
	accessTTL time.Duration,
	redirectURI string,
) AuthService {
	return &authService{
		repo:         repo,
		telegramOIDC: telegramOIDC,
		jwtSecret:    jwtSecret,
		accessTTL:    accessTTL,
		redirectURI:  redirectURI,
	}
}

func (s *authService) StartTelegramOAuth(ctx context.Context) (*domain.TelegramOAuthStart, string, string, string, error) {
	state, err := randomBase64URL(32)
	if err != nil {
		return nil, "", "", "", err
	}
	verifier, err := randomBase64URL(64)
	if err != nil {
		return nil, "", "", "", err
	}
	nonce, err := randomBase64URL(32)
	if err != nil {
		return nil, "", "", "", err
	}

	challenge := codeChallengeS256(verifier)
	authURL := s.telegramOIDC.BuildAuthURL(state, challenge, nonce)
	return &domain.TelegramOAuthStart{AuthURL: authURL, State: state}, state, verifier, nonce, nil
}

func (s *authService) FinishTelegramOAuth(ctx context.Context, input domain.TelegramOAuthCallbackInput) (*domain.AuthSession, error) {
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.State) == "" {
		return nil, domain.ErrInvalidInput
	}
	if input.ExpectedState == "" || input.CodeVerifier == "" {
		return nil, domain.ErrUnauthorized
	}
	if input.State != input.ExpectedState {
		return nil, domain.ErrUnauthorized
	}

	tokens, err := s.telegramOIDC.ExchangeCode(ctx, domain.TelegramOIDCTokenRequest{
		Code:         input.Code,
		RedirectURI:  s.redirectURI,
		CodeVerifier: input.CodeVerifier,
	})
	if err != nil {
		return nil, fmt.Errorf("finish telegram oauth: %w", err)
	}

	claims, err := s.telegramOIDC.VerifyIDToken(ctx, tokens.IDToken, input.Nonce)
	if err != nil {
		return nil, fmt.Errorf("verify telegram id token: %w", err)
	}

	var username *string
	if claims.PreferredUsername != "" {
		username = &claims.PreferredUsername
	}
	var firstName *string
	if claims.GivenName != "" {
		firstName = &claims.GivenName
	}
	var lastName *string
	if claims.FamilyName != "" {
		lastName = &claims.FamilyName
	}

	user, err := s.repo.Users.CreateOrUpdateTelegramUser(ctx, domain.TelegramUserInput{
		TelegramID:        claims.TelegramID,
		TelegramUsername:  username,
		TelegramFirstName: firstName,
		TelegramLastName:  lastName,
	})
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.repo.Admins.IsActiveByTelegramID(ctx, claims.TelegramID)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		_ = s.repo.Admins.MarkLogin(ctx, claims.TelegramID)
	}

	jwtClaims := appjwt.NewClaims(user.ID, user.TelegramID, isAdmin, s.accessTTL)
	accessToken, err := appjwt.Sign(s.jwtSecret, jwtClaims)
	if err != nil {
		return nil, err
	}

	return &domain.AuthSession{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   time.Unix(jwtClaims.ExpiresAt, 0),
		User:        user,
		IsAdmin:     isAdmin,
	}, nil
}

func randomBase64URL(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func codeChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
