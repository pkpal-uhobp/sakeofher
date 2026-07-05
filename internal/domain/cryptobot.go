package domain

import "time"

type CryptoBotCreateInvoiceRequest struct {
	Amount         string   `json:"amount"`
	CurrencyType   string   `json:"currency_type"`
	Fiat           string   `json:"fiat,omitempty"`
	AcceptedAssets []string `json:"accepted_assets,omitempty"`
	Description    string   `json:"description,omitempty"`
	Payload        string   `json:"payload,omitempty"`
	ExpiresIn      int      `json:"expires_in,omitempty"`
	PaidButtonName string   `json:"paid_btn_name,omitempty"`
	PaidButtonURL  string   `json:"paid_btn_url,omitempty"`
}

type CryptoBotInvoice struct {
	InvoiceID         int64      `json:"invoice_id"`
	Status            string     `json:"status"`
	Amount            string     `json:"amount"`
	Asset             string     `json:"asset"`
	Fiat              string     `json:"fiat"`
	BotInvoiceURL     string     `json:"bot_invoice_url"`
	MiniAppInvoiceURL string     `json:"mini_app_invoice_url"`
	WebAppInvoiceURL  string     `json:"web_app_invoice_url"`
	ExpirationDate    *time.Time `json:"expiration_date,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
}
