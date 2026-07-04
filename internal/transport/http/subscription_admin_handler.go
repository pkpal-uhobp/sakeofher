package httptransport

import (
	"context"
	"net/http"
	"strconv"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type SubscriptionAdminHandler struct {
	services *service.Services
}

func NewSubscriptionAdminHandler(services *service.Services) *SubscriptionAdminHandler {
	return &SubscriptionAdminHandler{services: services}
}

func (h *SubscriptionAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	telegramID, _ := strconv.ParseInt(r.URL.Query().Get("telegram_id"), 10, 64)

	input := domain.SubscriptionListInput{
		UserID:     userID,
		TelegramID: telegramID,
		Status:     domain.SubscriptionStatus(r.URL.Query().Get("status")),
		Limit:      limit,
		Offset:     offset,
	}

	out, err := h.services.Subscriptions.List(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *SubscriptionAdminHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	out, err := h.services.Subscriptions.GetByID(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *SubscriptionAdminHandler) CreateManual(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateManualSubscriptionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.services.Subscriptions.CreateManual(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, out)
}

func (h *SubscriptionAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	var input domain.UpdateSubscriptionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.services.Subscriptions.Update(r.Context(), id, input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *SubscriptionAdminHandler) Extend(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	var input domain.ExtendSubscriptionInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.services.Subscriptions.Extend(r.Context(), id, input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *SubscriptionAdminHandler) UpdateTrafficLimit(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	var input domain.UpdateTrafficLimitInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.services.Subscriptions.UpdateTrafficLimit(r.Context(), id, input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *SubscriptionAdminHandler) Disable(w http.ResponseWriter, r *http.Request) {
	h.change(w, r, h.services.Subscriptions.Disable)
}

func (h *SubscriptionAdminHandler) Enable(w http.ResponseWriter, r *http.Request) {
	h.change(w, r, h.services.Subscriptions.Enable)
}

func (h *SubscriptionAdminHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.change(w, r, h.services.Subscriptions.Cancel)
}

func (h *SubscriptionAdminHandler) change(
	w http.ResponseWriter,
	r *http.Request,
	fn func(rctx context.Context, id int64) (*domain.PublicSubscription, error),
) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	out, err := fn(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}
