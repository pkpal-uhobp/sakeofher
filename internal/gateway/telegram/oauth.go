package telegram

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sakeofher/internal/domain"
)

const (
	telegramOIDCIssuer = "https://oauth.telegram.org"
	telegramAuthURL    = "https://oauth.telegram.org/auth"
	telegramTokenURL   = "https://oauth.telegram.org/token"
	telegramJWKSURL    = "https://oauth.telegram.org/.well-known/jwks.json"
)

type OAuthClient struct {
	clientID     string
	clientSecret string
	redirectURI  string
	http         *http.Client

	mu        sync.Mutex
	jwks      *jwksResponse
	jwksUntil time.Time
}

func NewOAuthClient(clientID string, clientSecret string, redirectURI string, timeout time.Duration) *OAuthClient {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		http:         &http.Client{Timeout: timeout},
	}
}

func (c *OAuthClient) BuildAuthURL(state string, codeChallenge string, nonce string) string {
	q := url.Values{}
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", c.redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	if nonce != "" {
		q.Set("nonce", nonce)
	}
	return telegramAuthURL + "?" + q.Encode()
}

func (c *OAuthClient) ExchangeCode(ctx context.Context, req domain.TelegramOIDCTokenRequest) (*domain.TelegramOIDCTokenResponse, error) {
	if c.clientID == "" || c.clientSecret == "" || c.redirectURI == "" {
		return nil, fmt.Errorf("telegram oauth is not configured")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", req.Code)
	form.Set("redirect_uri", req.RedirectURI)
	form.Set("client_id", c.clientID)
	form.Set("code_verifier", req.CodeVerifier)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, telegramTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("telegram token exchange: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram token exchange status %d: %s", resp.StatusCode, string(body))
	}

	var out domain.TelegramOIDCTokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode telegram token response: %w", err)
	}
	if out.IDToken == "" {
		return nil, fmt.Errorf("telegram token response does not contain id_token")
	}
	return &out, nil
}

func (c *OAuthClient) VerifyIDToken(ctx context.Context, idToken string, expectedNonce string) (*domain.TelegramOIDCClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid telegram id_token format")
	}

	headerRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode telegram id_token header: %w", err)
	}
	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode telegram id_token payload: %w", err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode telegram id_token signature: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerRaw, &header); err != nil {
		return nil, fmt.Errorf("decode telegram id_token header json: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported telegram id_token alg: %s", header.Alg)
	}

	key, err := c.findRSAKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return nil, fmt.Errorf("invalid telegram id_token signature: %w", err)
	}

	var claims domain.TelegramOIDCClaims
	dec := json.NewDecoder(bytes.NewReader(payloadRaw))
	dec.UseNumber()
	if err := dec.Decode(&claims); err != nil {
		return nil, fmt.Errorf("decode telegram id_token claims: %w", err)
	}

	if claims.Issuer != telegramOIDCIssuer {
		return nil, fmt.Errorf("invalid telegram id_token issuer")
	}
	if !audienceContains(claims.Audience, c.clientID) {
		return nil, fmt.Errorf("invalid telegram id_token audience")
	}
	if claims.ExpiresAt <= time.Now().Unix() {
		return nil, fmt.Errorf("telegram id_token expired")
	}
	if expectedNonce != "" && subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(expectedNonce)) != 1 {
		return nil, fmt.Errorf("invalid telegram id_token nonce")
	}
	if claims.TelegramID <= 0 {
		return nil, fmt.Errorf("telegram id_token does not contain user id")
	}

	return &claims, nil
}

func (c *OAuthClient) findRSAKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	jwks, err := c.getJWKS(ctx)
	if err != nil {
		return nil, err
	}
	for _, k := range jwks.Keys {
		if k.Kid == kid && k.Kty == "RSA" {
			return k.rsaPublicKey()
		}
	}
	// One forced refresh in case Telegram rotated keys.
	jwks, err = c.fetchJWKS(ctx)
	if err != nil {
		return nil, err
	}
	for _, k := range jwks.Keys {
		if k.Kid == kid && k.Kty == "RSA" {
			return k.rsaPublicKey()
		}
	}
	return nil, fmt.Errorf("telegram jwks key not found: %s", kid)
}

func (c *OAuthClient) getJWKS(ctx context.Context) (*jwksResponse, error) {
	c.mu.Lock()
	if c.jwks != nil && time.Now().Before(c.jwksUntil) {
		jwks := c.jwks
		c.mu.Unlock()
		return jwks, nil
	}
	c.mu.Unlock()
	return c.fetchJWKS(ctx)
}

func (c *OAuthClient) fetchJWKS(ctx context.Context) (*jwksResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, telegramJWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch telegram jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch telegram jwks status: %d", resp.StatusCode)
	}
	var jwks jwksResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode telegram jwks: %w", err)
	}
	if len(jwks.Keys) == 0 {
		return nil, errors.New("telegram jwks is empty")
	}

	c.mu.Lock()
	c.jwks = &jwks
	c.jwksUntil = time.Now().Add(time.Hour)
	c.mu.Unlock()

	return &jwks, nil
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (k jwk) rsaPublicKey() (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode jwk n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode jwk e: %w", err)
	}
	if len(eBytes) == 0 {
		return nil, fmt.Errorf("empty jwk e")
	}

	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, fmt.Errorf("invalid jwk e")
	}

	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func audienceContains(aud any, expected string) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	case []string:
		for _, item := range v {
			if item == expected {
				return true
			}
		}
	}
	return false
}
