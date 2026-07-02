package remnawave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	// TODO: уточнить DTO Remnawave под фактическую версию API.
	payload, _ := json.Marshal(req)
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
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remnawave create user status: %d", resp.StatusCode)
	}
	return &domain.RemnaUser{Username: req.Username}, nil
}

func (c *Client) EnableUser(ctx context.Context, remnaUUID string) error   { return nil }
func (c *Client) DisableUser(ctx context.Context, remnaUUID string) error  { return nil }
func (c *Client) DeleteUser(ctx context.Context, remnaUUID string) error   { return nil }
func (c *Client) ResetTraffic(ctx context.Context, remnaUUID string) error { return nil }
func (c *Client) GetUserTraffic(ctx context.Context, remnaUUID string) (*domain.RemnaTraffic, error) {
	return &domain.RemnaTraffic{}, nil
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
}
