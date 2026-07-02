package tribute

import "time"

type Client struct {
	apiKey  string
	timeout time.Duration
}

func NewClient(apiKey string, timeout time.Duration) *Client {
	return &Client{apiKey: apiKey, timeout: timeout}
}
