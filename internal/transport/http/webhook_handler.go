package httptransport

import (
	"io"
	"net/http"
	"strings"

	"sakeofher/internal/service"
)

type WebhookHandler struct {
	services            *service.Services
	cryptoBotSecretPath string
}

func NewWebhookHandler(services *service.Services, cryptoBotSecretPath string) *WebhookHandler {
	return &WebhookHandler{
		services:            services,
		cryptoBotSecretPath: strings.Trim(strings.TrimSpace(cryptoBotSecretPath), "/"),
	}
}

func (h *WebhookHandler) CryptoBot(w http.ResponseWriter, r *http.Request) {
	secret := r.PathValue("secret")
	if h.cryptoBotSecretPath == "" || secret != h.cryptoBotSecretPath {
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}

	if err := h.services.Payments.HandleCryptoBotWebhook(r.Context(), raw); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}
