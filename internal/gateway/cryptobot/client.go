package cryptobot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

const defaultBaseURL = "https://pay.crypt.bot/api"

type Client struct {
	token   string
	timeout time.Duration
	baseURL string
	http    *http.Client
}

func NewClient(token string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &Client{
		token:   strings.TrimSpace(token),
		timeout: timeout,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) isStub() bool {
	return strings.TrimSpace(c.token) == ""
}

func (c *Client) CreateInvoice(ctx context.Context, req domain.CryptoBotCreateInvoiceRequest) (*domain.CryptoBotInvoice, error) {
	if c.isStub() {
		return nil, fmt.Errorf("cryptobot api token is empty")
	}

	payload := map[string]any{
		"currency_type": "fiat",
		"fiat":          "RUB",
		"amount":        strings.TrimSpace(req.Amount),
		"description":   strings.TrimSpace(req.Description),
		"payload":       strings.TrimSpace(req.Payload),
		"expires_in":    req.ExpiresIn,
	}

	assets := normalizeAssets(req.AcceptedAssets)
	if len(assets) > 0 {
		payload["accepted_assets"] = strings.Join(assets, ",")
	}

	if req.PaidButtonName != "" && req.PaidButtonURL != "" {
		payload["paid_btn_name"] = req.PaidButtonName
		payload["paid_btn_url"] = req.PaidButtonURL
	}

	var out cryptoBotInvoiceDTO
	if err := c.do(ctx, http.MethodPost, "/createInvoice", payload, &out); err != nil {
		return nil, err
	}

	return out.toDomain(), nil
}

func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (*domain.CryptoBotInvoice, error) {
	if c.isStub() {
		return nil, fmt.Errorf("cryptobot api token is empty")
	}

	id, err := strconv.ParseInt(strings.TrimSpace(invoiceID), 10, 64)
	if err != nil || id <= 0 {
		return nil, domain.ErrInvalidInput
	}

	var out struct {
		Items []cryptoBotInvoiceDTO `json:"items"`
	}

	payload := map[string]any{
		"invoice_ids": strconv.FormatInt(id, 10),
	}

	if err := c.do(ctx, http.MethodPost, "/getInvoices", payload, &out); err != nil {
		return nil, err
	}

	if len(out.Items) == 0 {
		return nil, domain.ErrNotFound
	}

	return out.Items[0].toDomain(), nil
}

func (c *Client) do(ctx context.Context, method string, path string, payload any, result any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.baseURL, "/")+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Crypto-Pay-API-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cryptobot request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("cryptobot status %d: %s", resp.StatusCode, string(raw))
	}

	var wrapper struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return fmt.Errorf("cryptobot decode response: %w", err)
	}

	if !wrapper.OK {
		if wrapper.Error == "" {
			wrapper.Error = string(raw)
		}
		return fmt.Errorf("cryptobot api error: %s", wrapper.Error)
	}

	if result != nil && len(wrapper.Result) > 0 {
		if err := json.Unmarshal(wrapper.Result, result); err != nil {
			return fmt.Errorf("cryptobot decode result: %w", err)
		}
	}

	return nil
}

type cryptoBotInvoiceDTO struct {
	InvoiceID         int64  `json:"invoice_id"`
	Status            string `json:"status"`
	Amount            string `json:"amount"`
	Asset             string `json:"asset"`
	Fiat              string `json:"fiat"`
	BotInvoiceURL     string `json:"bot_invoice_url"`
	MiniAppInvoiceURL string `json:"mini_app_invoice_url"`
	WebAppInvoiceURL  string `json:"web_app_invoice_url"`
	ExpirationDate    any    `json:"expiration_date"`
	PaidAt            any    `json:"paid_at"`
}

func (d cryptoBotInvoiceDTO) toDomain() *domain.CryptoBotInvoice {
	return &domain.CryptoBotInvoice{
		InvoiceID:         d.InvoiceID,
		Status:            strings.ToLower(strings.TrimSpace(d.Status)),
		Amount:            d.Amount,
		Asset:             d.Asset,
		Fiat:              d.Fiat,
		BotInvoiceURL:     d.BotInvoiceURL,
		MiniAppInvoiceURL: d.MiniAppInvoiceURL,
		WebAppInvoiceURL:  d.WebAppInvoiceURL,
		ExpirationDate:    parseCryptoBotTime(d.ExpirationDate),
		PaidAt:            parseCryptoBotTime(d.PaidAt),
	}
}

func parseCryptoBotTime(v any) *time.Time {
	switch value := v.(type) {
	case float64:
		if value <= 0 {
			return nil
		}
		t := time.Unix(int64(value), 0)
		return &t

	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return nil
		}

		if unix, err := strconv.ParseInt(value, 10, 64); err == nil && unix > 0 {
			t := time.Unix(unix, 0)
			return &t
		}

		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return &t
		}
	}

	return nil
}

func normalizeAssets(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		item = strings.ToUpper(strings.TrimSpace(item))
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
	}

	return out
}
