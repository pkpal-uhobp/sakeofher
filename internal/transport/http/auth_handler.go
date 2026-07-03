package httptransport

import (
	"net/http"
	"strings"
	"time"

	"sakeofher/internal/domain"
	appjwt "sakeofher/internal/platform/jwt"
	"sakeofher/internal/service"
)

const (
	telegramOAuthStateCookie    = "sakeofher_tg_oauth_state"
	telegramOAuthVerifierCookie = "sakeofher_tg_oauth_verifier"
	telegramOAuthNonceCookie    = "sakeofher_tg_oauth_nonce"
	accessTokenCookie           = "sakeofher_access_token"
)

type AuthHandler struct {
	services           *service.Services
	successRedirectURL string
	jwtSecret          string
}

func NewAuthHandler(services *service.Services, successRedirectURL string, jwtSecret string) *AuthHandler {
	return &AuthHandler{services: services, successRedirectURL: successRedirectURL, jwtSecret: jwtSecret}
}

func (h *AuthHandler) TelegramOAuthStart(w http.ResponseWriter, r *http.Request) {
	start, state, verifier, nonce, err := h.services.Auth.StartTelegramOAuth(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	secure := isHTTPS(r)
	setTempCookie(w, telegramOAuthStateCookie, state, secure)
	setTempCookie(w, telegramOAuthVerifierCookie, verifier, secure)
	setTempCookie(w, telegramOAuthNonceCookie, nonce, secure)

	if wantsJSON(r) {
		WriteJSON(w, http.StatusOK, start)
		return
	}
	http.Redirect(w, r, start.AuthURL, http.StatusFound)
}

func (h *AuthHandler) TelegramOAuthURL(w http.ResponseWriter, r *http.Request) {
	start, state, verifier, nonce, err := h.services.Auth.StartTelegramOAuth(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	secure := isHTTPS(r)
	setTempCookie(w, telegramOAuthStateCookie, state, secure)
	setTempCookie(w, telegramOAuthVerifierCookie, verifier, secure)
	setTempCookie(w, telegramOAuthNonceCookie, nonce, secure)

	WriteJSON(w, http.StatusOK, start)
}

func (h *AuthHandler) TelegramOAuthCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(telegramOAuthStateCookie)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "telegram oauth state cookie is missing")
		return
	}
	verifierCookie, err := r.Cookie(telegramOAuthVerifierCookie)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "telegram oauth verifier cookie is missing")
		return
	}
	nonceCookie, err := r.Cookie(telegramOAuthNonceCookie)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "telegram oauth nonce cookie is missing")
		return
	}

	input := domain.TelegramOAuthCallbackInput{
		Code:          r.URL.Query().Get("code"),
		State:         r.URL.Query().Get("state"),
		ExpectedState: stateCookie.Value,
		CodeVerifier:  verifierCookie.Value,
		Nonce:         nonceCookie.Value,
	}

	session, err := h.services.Auth.FinishTelegramOAuth(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	secure := isHTTPS(r)
	clearCookie(w, telegramOAuthStateCookie, secure)
	clearCookie(w, telegramOAuthVerifierCookie, secure)
	clearCookie(w, telegramOAuthNonceCookie, secure)
	setAccessCookie(w, session.AccessToken, time.Until(session.ExpiresAt), secure)

	if wantsJSON(r) || h.successRedirectURL == "" {
		WriteJSON(w, http.StatusOK, session)
		return
	}

	http.Redirect(w, r, h.successRedirectURL, http.StatusFound)
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

	user, err := h.services.Users.GetByTelegramID(r.Context(), claims.TelegramID)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"user":     user,
		"is_admin": claims.IsAdmin,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, accessTokenCookie, isHTTPS(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func setTempCookie(w http.ResponseWriter, name string, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
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

func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || r.URL.Query().Get("format") == "json"
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
