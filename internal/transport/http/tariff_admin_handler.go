package httptransport

import (
	"context"
	"net/http"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type TariffAdminHandler struct {
	services *service.Services
}

func NewTariffAdminHandler(services *service.Services) *TariffAdminHandler {
	return &TariffAdminHandler{services: services}
}

func (h *TariffAdminHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.Tariffs.ListAll(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, items)
}

func (h *TariffAdminHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	item, err := h.services.Tariffs.GetByID(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, item)
}

func (h *TariffAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateTariffInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Tariffs.Create(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, item)
}

func (h *TariffAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	var input domain.UpdateTariffInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Tariffs.Update(r.Context(), id, input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, item)
}

func (h *TariffAdminHandler) Enable(w http.ResponseWriter, r *http.Request) {
	h.change(w, r, h.services.Tariffs.Enable)
}

func (h *TariffAdminHandler) Disable(w http.ResponseWriter, r *http.Request) {
	h.change(w, r, h.services.Tariffs.Disable)
}

func (h *TariffAdminHandler) change(
	w http.ResponseWriter,
	r *http.Request,
	fn func(rctx context.Context, id int64) (*domain.Tariff, error),
) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	item, err := fn(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, item)
}
