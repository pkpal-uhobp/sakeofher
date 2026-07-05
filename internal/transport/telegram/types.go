package telegramtransport

type update struct {
	UpdateID         int               `json:"update_id"`
	Message          *message          `json:"message,omitempty"`
	CallbackQuery    *callbackQuery    `json:"callback_query,omitempty"`
	PreCheckoutQuery *preCheckoutQuery `json:"pre_checkout_query,omitempty"`
}

type message struct {
	MessageID         int                `json:"message_id"`
	From              *tgUser            `json:"from,omitempty"`
	Chat              chat               `json:"chat"`
	Text              string             `json:"text,omitempty"`
	SuccessfulPayment *successfulPayment `json:"successful_payment,omitempty"`
}

type callbackQuery struct {
	ID      string   `json:"id"`
	From    tgUser   `json:"from"`
	Message *message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

type preCheckoutQuery struct {
	ID             string `json:"id"`
	From           tgUser `json:"from"`
	Currency       string `json:"currency"`
	TotalAmount    int64  `json:"total_amount"`
	InvoicePayload string `json:"invoice_payload"`
}

type successfulPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int64  `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string `json:"provider_payment_charge_id"`
}

type tgUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type botCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type inlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
	Pay          bool   `json:"pay,omitempty"`
}

type labeledPrice struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

type starBalance struct {
	Amount         int64 `json:"amount"`
	NanostarAmount int64 `json:"nanostar_amount,omitempty"`
}
