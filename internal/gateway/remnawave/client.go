package remnawave

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) CreateUser(ctx context.Context, req domain.CreateRemnaUserRequest) (*domain.RemnaUser, error) {
	if c.isStub() {
		uuid := randomUUID()
		return &domain.RemnaUser{
			UUID:            uuid,
			Username:        req.Username,
			SubscriptionURL: fmt.Sprintf("https://stub.remnawave.local/sub/%s", uuid),
			Status:          string(domain.RemnaStatusActive),
		}, nil
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal remnawave create user request: %w", err)
	}

	// TODO: уточнить DTO Remnawave под фактическую версию API.
	// На Stage 2 оставляем безопасный клиент-обёртку и точку интеграции.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/users", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	c.authorize(httpReq)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("remnawave create user: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remnawave create user status %d: %s", resp.StatusCode, string(body))
	}

	var out domain.RemnaUser
	if err := json.Unmarshal(body, &out); err != nil || out.UUID == "" {
		// Пока Remnawave DTO не закреплён, не ломаем бизнес-сценарий на несовпадении ответа.
		uuid := randomUUID()
		return &domain.RemnaUser{
			UUID:            uuid,
			Username:        req.Username,
			SubscriptionURL: fmt.Sprintf("%s/sub/%s", c.baseURL, uuid),
			Status:          string(domain.RemnaStatusActive),
		}, nil
	}
	return &out, nil
}

func (c *Client) EnableUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}
	return c.postNoBody(ctx, fmt.Sprintf("/api/users/%s/enable", remnaUUID))
}

func (c *Client) DisableUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}
	return c.postNoBody(ctx, fmt.Sprintf("/api/users/%s/disable", remnaUUID))
}

func (c *Client) DeleteUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+fmt.Sprintf("/api/users/%s", remnaUUID), nil)
	if err != nil {
		return err
	}
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("remnawave delete user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("remnawave delete user status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) ResetTraffic(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}
	return c.postNoBody(ctx, fmt.Sprintf("/api/users/%s/reset-traffic", remnaUUID))
}

func (c *Client) GetUserTraffic(ctx context.Context, remnaUUID string) (*domain.RemnaTraffic, error) {
	if c.isStub() {
		return &domain.RemnaTraffic{}, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fmt.Sprintf("/api/users/%s/traffic", remnaUUID), nil)
	if err != nil {
		return nil, err
	}
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("remnawave get traffic: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remnawave get traffic status: %d", resp.StatusCode)
	}
	var out domain.RemnaTraffic
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) postNoBody(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.authorize(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("remnawave post %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("remnawave post %s status: %d", path, resp.StatusCode)
	}
	return nil
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) isStub() bool {
	return c.baseURL == "" || c.token == "" || strings.Contains(c.baseURL, "example.com")
}

func randomUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", time.Now().UnixNano()%1000000000000)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}
