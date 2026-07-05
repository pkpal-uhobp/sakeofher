package telegramtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"sakeofher/internal/service"
)

const telegramAPIBaseURL = "https://api.telegram.org/bot"

type Bot struct {
	token    string
	apiURL   string
	client   *http.Client
	services *service.Services
	log      *zap.Logger
	settings botSettings
	router   *Router

	stateMu sync.RWMutex
	states  map[int64]adminState

	broadcastMu sync.RWMutex
	broadcasts  map[int64]broadcastDraft
}

func NewBot(token string, services *service.Services, log *zap.Logger) *Bot {
	b := &Bot{
		token:      strings.TrimSpace(token),
		client:     &http.Client{Timeout: 35 * time.Second},
		services:   services,
		log:        log,
		settings:   loadBotSettings(),
		states:     make(map[int64]adminState),
		broadcasts: make(map[int64]broadcastDraft),
	}
	b.apiURL = telegramAPIBaseURL + b.token + "/"
	b.router = NewRouter(b)
	return b
}

func (b *Bot) Run(ctx context.Context) error {
	if b.token == "" {
		return errors.New("telegram bot token is empty: set TELEGRAM_BOT_TOKEN")
	}
	if b.services == nil || b.services.Users == nil || b.services.Tariffs == nil || b.services.Subscriptions == nil || b.services.Payments == nil {
		return errors.New("telegram bot services are not initialized")
	}

	b.log.Info("telegram bot started", zap.String("mode", "long_polling"))
	if err := b.deleteWebhook(ctx); err != nil {
		b.log.Warn("telegram deleteWebhook failed", zap.Error(err))
	}
	if err := b.setCommands(ctx); err != nil {
		b.log.Warn("telegram setMyCommands failed", zap.Error(err))
	}

	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			b.log.Warn("telegram getUpdates failed", zap.Error(err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
				continue
			}
		}

		for _, upd := range updates {
			if upd.UpdateID >= offset {
				offset = upd.UpdateID + 1
			}
			if err := b.router.Handle(ctx, upd); err != nil {
				b.log.Warn("telegram update handling failed", zap.Int("update_id", upd.UpdateID), zap.Error(err))
			}
		}
	}
}

func (b *Bot) getUpdates(ctx context.Context, offset int) ([]update, error) {
	var result []update
	payload := map[string]any{
		"offset":          offset,
		"timeout":         25,
		"allowed_updates": []string{"message", "callback_query", "pre_checkout_query"},
	}
	if err := b.call(ctx, "getUpdates", payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *Bot) deleteWebhook(ctx context.Context) error {
	var ok bool
	return b.call(ctx, "deleteWebhook", map[string]any{"drop_pending_updates": false}, &ok)
}

func (b *Bot) setCommands(ctx context.Context) error {
	commands := []botCommand{
		{Command: "start", Description: "Главное меню"},
		{Command: "status", Description: "Моя подписка"},
		{Command: "renew", Description: "Продлить подписку"},
		{Command: "help", Description: "Помощь"},
		{Command: "admin", Description: "Админ-панель"},
		{Command: "grant", Description: "Выдать или продлить подписку"},
		{Command: "check", Description: "Проверить пользователя"},
		{Command: "extend", Description: "Продлить на дни"},
		{Command: "traffic", Description: "Изменить лимит трафика"},
		{Command: "enable", Description: "Включить подписку"},
		{Command: "disable", Description: "Отключить подписку"},
		{Command: "users", Description: "Список пользователей"},
		{Command: "subs", Description: "Список подписок"},
		{Command: "stars", Description: "Баланс звезд"},
		{Command: "broadcast", Description: "Рассылка всем активным пользователям"},
	}
	var ok bool
	return b.call(ctx, "setMyCommands", map[string]any{"commands": commands}, &ok)
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string, markup *inlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
	}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	var msg message
	return b.call(ctx, "sendMessage", payload, &msg)
}

func (b *Bot) editMessageText(ctx context.Context, chatID int64, messageID int, text string, markup *inlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":                  chatID,
		"message_id":               messageID,
		"text":                     text,
		"disable_web_page_preview": true,
	}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	var result any
	return b.call(ctx, "editMessageText", payload, &result)
}

func (b *Bot) answerCallback(ctx context.Context, callbackID string, text string, alert bool) error {
	payload := map[string]any{
		"callback_query_id": callbackID,
		"text":              text,
		"show_alert":        alert,
		"cache_time":        0,
	}
	var ok bool
	return b.call(ctx, "answerCallbackQuery", payload, &ok)
}

func (b *Bot) answerPreCheckout(ctx context.Context, queryID string, ok bool, errorMessage string) error {
	payload := map[string]any{
		"pre_checkout_query_id": queryID,
		"ok":                    ok,
	}
	if !ok {
		payload["error_message"] = errorMessage
	}
	var result bool
	return b.call(ctx, "answerPreCheckoutQuery", payload, &result)
}

func (b *Bot) sendInvoice(ctx context.Context, chatID int64, title string, description string, payload string, starsAmount int64) error {
	invoiceKeyboard := &inlineKeyboardMarkup{InlineKeyboard: [][]inlineKeyboardButton{
		{{Text: fmt.Sprintf("Оплатить %d ", starsAmount), Pay: true}},
	}}
	req := map[string]any{
		"chat_id":        chatID,
		"title":          title,
		"description":    description,
		"payload":        payload,
		"provider_token": "",
		"currency":       "XTR",
		"prices":         []labeledPrice{{Label: title, Amount: starsAmount}},
		"reply_markup":   invoiceKeyboard,
	}
	var msg message
	return b.call(ctx, "sendInvoice", req, &msg)
}

func (b *Bot) getStarBalance(ctx context.Context) (*starBalance, error) {
	var result starBalance
	if err := b.call(ctx, "getMyStarBalance", map[string]any{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (b *Bot) replyOrEdit(ctx context.Context, upd update, text string, markup *inlineKeyboardMarkup) error {
	if upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil {
		chatID := upd.CallbackQuery.Message.Chat.ID
		messageID := upd.CallbackQuery.Message.MessageID
		if err := b.editMessageText(ctx, chatID, messageID, text, markup); err == nil {
			return nil
		}
		return b.sendMessage(ctx, chatID, text, markup)
	}
	if upd.Message != nil {
		return b.sendMessage(ctx, upd.Message.Chat.ID, text, markup)
	}
	return nil
}

func (b *Bot) call(ctx context.Context, method string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram %s request: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.apiURL+method, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create telegram %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram %s request: %w", method, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram %s response: %w", method, err)
	}

	var envelope struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("decode telegram %s response: %w; raw=%s", method, err, string(raw))
	}
	if !envelope.OK || resp.StatusCode >= 400 {
		if envelope.Description == "" {
			envelope.Description = resp.Status
		}
		return fmt.Errorf("telegram %s failed: %s", method, envelope.Description)
	}
	if result != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, result); err != nil {
			return fmt.Errorf("decode telegram %s result: %w", method, err)
		}
	}
	return nil
}
