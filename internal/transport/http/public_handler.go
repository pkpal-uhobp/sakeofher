package httptransport

import (
	"net/http"

	"sakeofher/internal/service"
)

type PublicHandler struct{ services *service.Services }

func NewPublicHandler(services *service.Services) *PublicHandler {
	return &PublicHandler{services: services}
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
