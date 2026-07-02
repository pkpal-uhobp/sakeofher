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
	tariffs, err := h.services.Tariffs.ListActive(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tariffs)
}
