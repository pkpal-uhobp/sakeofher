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
	"net/url"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

const (
	defaultTrafficResetStrategy = "NO_RESET"
	statusActive                = "ACTIVE"
	statusDisabled              = "DISABLED"

	noHWIDDeviceLimit = 0
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

type createUserRequestDTO struct {
	Username             string   `json:"username"`
	Status               string   `json:"status,omitempty"`
	TrafficLimitBytes    int64    `json:"trafficLimitBytes,omitempty"`
	TrafficLimitStrategy string   `json:"trafficLimitStrategy,omitempty"`
	ExpireAt             string   `json:"expireAt"`
	Description          string   `json:"description,omitempty"`
	TelegramID           *int64   `json:"telegramId,omitempty"`
	Email                *string  `json:"email,omitempty"`
	Tag                  *string  `json:"tag,omitempty"`
	HWIDDeviceLimit      int      `json:"hwidDeviceLimit"`
	ActiveInternalSquads []string `json:"activeInternalSquads,omitempty"`
}

type updateUserRequestDTO struct {
	UUID                 string   `json:"uuid"`
	Username             string   `json:"username,omitempty"`
	Status               string   `json:"status,omitempty"`
	TrafficLimitBytes    *int64   `json:"trafficLimitBytes,omitempty"`
	TrafficLimitStrategy string   `json:"trafficLimitStrategy,omitempty"`
	ExpireAt             *string  `json:"expireAt,omitempty"`
	Description          *string  `json:"description,omitempty"`
	TelegramID           *int64   `json:"telegramId,omitempty"`
	Email                *string  `json:"email,omitempty"`
	Tag                  *string  `json:"tag,omitempty"`
	HWIDDeviceLimit      int      `json:"hwidDeviceLimit"`

	// Do NOT add omitempty.
	// [] must be sent to Remnawave to remove user from all internal squads.
	ActiveInternalSquads []string `json:"activeInternalSquads"`
}

type userResponseDTO struct {
	UUID                     string     `json:"uuid"`
	ShortUUID                string     `json:"shortUuid"`
	Username                 string     `json:"username"`
	Status                   string     `json:"status"`
	UsedTrafficBytes         int64      `json:"usedTrafficBytes"`
	LifetimeUsedTrafficBytes int64      `json:"lifetimeUsedTrafficBytes"`
	TrafficLimitBytes        int64      `json:"trafficLimitBytes"`
	TrafficLimitStrategy     string     `json:"trafficLimitStrategy"`
	ExpireAt                 *time.Time `json:"expireAt"`
	LastTrafficResetAt        *time.Time `json:"lastTrafficResetAt"`
	SubscriptionURL          string     `json:"subscriptionUrl"`
}

type wrappedResponse[T any] struct {
	Response T `json:"response"`
}

type remnaError struct {
	StatusCode int    `json:"statusCode"`
	Message    any    `json:"message"`
	Error      string `json:"error"`
	ErrorCode  string `json:"errorCode"`
	Path       string `json:"path"`
}

func (c *Client) CreateUser(ctx context.Context, req domain.CreateRemnaUserRequest) (*domain.RemnaUser, error) {
	if c.isStub() {
		return stubUser(req.Username, req.TrafficLimitBytes, req.ExpiresAtUnix), nil
	}

	if strings.TrimSpace(req.Username) == "" || req.ExpiresAtUnix <= 0 {
		return nil, domain.ErrInvalidInput
	}

	strategy := req.TrafficResetStrategy
	if strategy == "" {
		strategy = defaultTrafficResetStrategy
	}

	dto := createUserRequestDTO{
		Username:             req.Username,
		Status:               statusActive,
		TrafficLimitBytes:    req.TrafficLimitBytes,
		TrafficLimitStrategy: strategy,
		ExpireAt:             time.Unix(req.ExpiresAtUnix, 0).UTC().Format(time.RFC3339Nano),
		Description:          req.Description,
		TelegramID:           req.TelegramID,
		Email:                req.Email,
		Tag:                  req.Tag,
		HWIDDeviceLimit:      noHWIDDeviceLimit,
		ActiveInternalSquads: req.ActiveInternalSquads,
	}

	var out userResponseDTO
	if err := c.doJSON(ctx, http.MethodPost, "/api/users", dto, &out); err != nil {
		return nil, fmt.Errorf("remnawave create user: %w", err)
	}

	return mapUser(out), nil
}

func (c *Client) UpdateUser(ctx context.Context, req domain.UpdateRemnaUserRequest) (*domain.RemnaUser, error) {
	if c.isStub() {
		return stubUser(req.Username, valueInt64(req.TrafficLimitBytes), valueUnix(req.ExpiresAtUnix)), nil
	}

	if strings.TrimSpace(req.UUID) == "" {
		return nil, domain.ErrInvalidInput
	}

	var expireAt *string
	if req.ExpiresAtUnix != nil && *req.ExpiresAtUnix > 0 {
		value := time.Unix(*req.ExpiresAtUnix, 0).UTC().Format(time.RFC3339Nano)
		expireAt = &value
	}

	strategy := req.TrafficResetStrategy
	if strategy == "" && req.TrafficLimitBytes != nil {
		strategy = defaultTrafficResetStrategy
	}

	dto := updateUserRequestDTO{
		UUID:                 req.UUID,
		Username:             req.Username,
		Status:               req.Status,
		TrafficLimitBytes:    req.TrafficLimitBytes,
		TrafficLimitStrategy: strategy,
		ExpireAt:             expireAt,
		Description:          req.Description,
		TelegramID:           req.TelegramID,
		Email:                req.Email,
		Tag:                  req.Tag,
		HWIDDeviceLimit:      noHWIDDeviceLimit,
		ActiveInternalSquads: req.ActiveInternalSquads,
	}

	var out userResponseDTO
	if err := c.doJSON(ctx, http.MethodPatch, "/api/users", dto, &out); err != nil {
		return nil, fmt.Errorf("remnawave update user: %w", err)
	}

	return mapUser(out), nil
}

func (c *Client) GetUser(ctx context.Context, remnaUUID string) (*domain.RemnaUser, error) {
	if c.isStub() {
		return &domain.RemnaUser{
			UUID:              remnaUUID,
			Username:          "stub",
			Status:            statusActive,
			TrafficLimitBytes: 0,
			UsedTrafficBytes:  0,
		}, nil
	}

	if strings.TrimSpace(remnaUUID) == "" {
		return nil, domain.ErrInvalidInput
	}

	var out userResponseDTO
	if err := c.doJSON(ctx, http.MethodGet, "/api/users/"+url.PathEscape(remnaUUID), nil, &out); err != nil {
		return nil, fmt.Errorf("remnawave get user: %w", err)
	}

	return mapUser(out), nil
}

func (c *Client) EnableUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}

	return c.doNoResponse(ctx, http.MethodPost, "/api/users/"+url.PathEscape(remnaUUID)+"/actions/enable", nil)
}

func (c *Client) DisableUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}

	return c.doNoResponse(ctx, http.MethodPost, "/api/users/"+url.PathEscape(remnaUUID)+"/actions/disable", nil)
}

func (c *Client) DeleteUser(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}

	return c.doNoResponse(ctx, http.MethodDelete, "/api/users/"+url.PathEscape(remnaUUID), nil)
}

func (c *Client) ResetTraffic(ctx context.Context, remnaUUID string) error {
	if c.isStub() {
		return nil
	}

	return c.doNoResponse(ctx, http.MethodPost, "/api/users/"+url.PathEscape(remnaUUID)+"/actions/reset-traffic", nil)
}

func (c *Client) GetUserTraffic(ctx context.Context, remnaUUID string) (*domain.RemnaTraffic, error) {
	user, err := c.GetUser(ctx, remnaUUID)
	if err != nil {
		return nil, err
	}

	return &domain.RemnaTraffic{
		UsedBytes:  user.UsedTrafficBytes,
		LimitBytes: user.TrafficLimitBytes,
	}, nil
}

func (c *Client) doNoResponse(ctx context.Context, method string, path string, payload any) error {
	return c.doJSON(ctx, method, path, payload, nil)
}

func (c *Client) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	var body io.Reader

	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}

	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))

	if resp.StatusCode >= 300 {
		var apiErr remnaError
		if err := json.Unmarshal(raw, &apiErr); err == nil && (apiErr.Error != "" || apiErr.Message != nil || apiErr.ErrorCode != "") {
			return fmt.Errorf("status %d: %+v", resp.StatusCode, apiErr)
		}

		return fmt.Errorf("status %d: %s", resp.StatusCode, string(raw))
	}

	if out == nil {
		return nil
	}

	var wrapped wrappedResponse[json.RawMessage]
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Response) > 0 {
		if err := json.Unmarshal(wrapped.Response, out); err != nil {
			return fmt.Errorf("decode wrapped response: %w", err)
		}

		return nil
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) authorize(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
}

func (c *Client) isStub() bool {
	return c.baseURL == "" || c.token == "" || strings.Contains(c.baseURL, "example.com")
}

func mapUser(user userResponseDTO) *domain.RemnaUser {
	status := user.Status
	if status == "" {
		status = statusActive
	}

	return &domain.RemnaUser{
		UUID:                     user.UUID,
		ShortUUID:                user.ShortUUID,
		Username:                 user.Username,
		SubscriptionURL:          user.SubscriptionURL,
		Status:                   status,
		UsedTrafficBytes:         user.UsedTrafficBytes,
		LifetimeUsedTrafficBytes: user.LifetimeUsedTrafficBytes,
		TrafficLimitBytes:        user.TrafficLimitBytes,
		TrafficLimitStrategy:     user.TrafficLimitStrategy,
		ExpireAt:                 user.ExpireAt,
		LastTrafficResetAt:        user.LastTrafficResetAt,
	}
}

func stubUser(username string, trafficLimitBytes int64, expiresAtUnix int64) *domain.RemnaUser {
	uuid := randomUUID()
	expireAt := time.Unix(expiresAtUnix, 0).UTC()

	return &domain.RemnaUser{
		UUID:                 uuid,
		ShortUUID:            strings.Split(uuid, "-")[0],
		Username:             username,
		SubscriptionURL:      fmt.Sprintf("https://stub.remnawave.local/sub/%s", uuid),
		Status:               statusActive,
		TrafficLimitBytes:    trafficLimitBytes,
		TrafficLimitStrategy: defaultTrafficResetStrategy,
		ExpireAt:             &expireAt,
	}
}

func valueInt64(value *int64) int64 {
	if value == nil {
		return 0
	}

	return *value
}

func valueUnix(value *int64) int64 {
	if value == nil {
		return time.Now().AddDate(0, 0, 30).Unix()
	}

	return *value
}

func randomUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", time.Now().UnixNano()%1000000000000)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
