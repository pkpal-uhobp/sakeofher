package httptransport

import (
	"net/http"
	"strings"
	"time"

	"sakeofher/internal/domain"
	appjwt "sakeofher/internal/platform/jwt"
	"sakeofher/internal/service"
)

const accessTokenCookie = "sakeofher_access_token"

type AuthHandler struct {
	services  *service.Services
	jwtSecret string
}

func NewAuthHandler(services *service.Services, jwtSecret string) *AuthHandler {
	return &AuthHandler{services: services, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input domain.LoginInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	session, err := h.services.Auth.Login(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	setAccessCookie(w, session.AccessToken, time.Until(session.ExpiresAt), isHTTPS(r))
	WriteJSON(w, http.StatusOK, session)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		if cookie, err := r.Cookie(accessTokenCookie); err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		WriteError(w, http.StatusUnauthorized, "missing access token")
		return
	}

	claims, err := appjwt.Verify(h.jwtSecret, token)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid access token")
		return
	}

	username := claims.Username
	if username == "" {
		username = claims.Subject
	}
	if username == "" {
		username = "admin"
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"username": username,
		"is_admin": claims.IsAdmin,
		"user": map[string]any{
			"id":                   1,
			"telegram_id":          0,
			"telegram_username":    username,
			"telegram_first_name":  username,
			"telegram_last_name":   "",
			"status":               "active",
			"remna_status":         "not_created",
			"subscription_url":      nil,
			"remna_uuid":           nil,
			"created_at":           time.Now(),
			"updated_at":           time.Now(),
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, accessTokenCookie, isHTTPS(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func setAccessCookie(w http.ResponseWriter, token string, ttl time.Duration, secure bool) {
	maxAge := int(ttl.Seconds())
	if maxAge <= 0 {
		maxAge = 1
	}

	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	proto := r.Header.Get("X-Forwarded-Proto")
	return strings.EqualFold(proto, "https")
}

func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
