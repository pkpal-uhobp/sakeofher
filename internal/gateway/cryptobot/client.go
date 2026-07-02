package cryptobot

import "time"

type Client struct {
	token   string
	timeout time.Duration
}

func NewClient(token string, timeout time.Duration) *Client {
	return &Client{token: token, timeout: timeout}
}
