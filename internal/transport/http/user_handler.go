package httptransport

import (
	"net/http"
	"strconv"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type BotHandler struct{ services *service.Services }

func NewBotHandler(services *service.Services) *BotHandler { return &BotHandler{services: services} }

func (h *BotHandler) Start(w http.ResponseWriter, r *http.Request) {
	var input domain.TelegramUserInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.services.Users.GetOrCreateTelegramUser(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (h *BotHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		WriteError(w, http.StatusBadRequest, "invalid telegram_id")
		return
	}
	item, err := h.services.Subscriptions.GetActiveByTelegramID(r.Context(), telegramID)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
