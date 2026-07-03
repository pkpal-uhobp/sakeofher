package httptransport

import (
	"net/http"
	"strconv"
	"strings"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type PublicHandler struct {
	services               *service.Services
	subscriptionPathSecret string
}

func NewPublicHandler(services *service.Services, subscriptionPathSecret string) *PublicHandler {
	return &PublicHandler{
		services:               services,
		subscriptionPathSecret: strings.Trim(strings.TrimSpace(subscriptionPathSecret), "/"),
	}
}

func (h *PublicHandler) ListTariffs(w http.ResponseWriter, r *http.Request) {
	tariffs, err := h.services.Tariffs.ListActiveWithPrices(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, tariffs)
}

func (h *PublicHandler) GetPublicSubscription(w http.ResponseWriter, r *http.Request) {
	item, err := h.services.Subscriptions.GetPublicByToken(r.Context(), r.PathValue("public_token"))
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicHandler) GetPublicSubscriptionByTelegramID(w http.ResponseWriter, r *http.Request) {
	if !h.isValidSubscriptionPath(r.PathValue("subscription_path")) {
		WriteDomainError(w, domain.ErrNotFound)
		return
	}

	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		WriteDomainError(w, domain.ErrInvalidInput)
		return
	}

	item, err := h.services.Subscriptions.GetLatestByTelegramID(r.Context(), telegramID)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *PublicHandler) isValidSubscriptionPath(path string) bool {
	path = strings.Trim(strings.TrimSpace(path), "/")
	return h.subscriptionPathSecret != "" && path == h.subscriptionPathSecret
}
