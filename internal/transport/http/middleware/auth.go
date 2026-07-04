package middleware

import (
	"context"
	"net/http"
	"strings"

	appjwt "sakeofher/internal/platform/jwt"
)

const accessTokenCookie = "sakeofher_access_token"

type authContextKey struct{}

func ClaimsFromContext(ctx context.Context) (*appjwt.Claims, bool) {
	claims, ok := ctx.Value(authContextKey{}).(*appjwt.Claims)
	return claims, ok
}

func RequireAuth(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authenticateRequest(jwtSecret, r)
		if !ok {
			writeAuthError(w, http.StatusUnauthorized, "missing or invalid access token")
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAdmin(jwtSecret string, next http.Handler) http.Handler {
	return RequireAuth(jwtSecret, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || !claims.IsAdmin {
			writeAuthError(w, http.StatusForbidden, "admin access required")
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func OptionalAuth(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authenticateRequest(jwtSecret, r)
		if ok {
			ctx := context.WithValue(r.Context(), authContextKey{}, claims)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func authenticateRequest(jwtSecret string, r *http.Request) (*appjwt.Claims, bool) {
	token := bearerToken(r)
	if token == "" {
		if cookie, err := r.Cookie(accessTokenCookie); err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		return nil, false
	}

	claims, err := appjwt.Verify(jwtSecret, token)
	if err != nil {
		return nil, false
	}

	return claims, true
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

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
