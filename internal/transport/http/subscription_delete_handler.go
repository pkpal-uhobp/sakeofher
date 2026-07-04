package httptransport

import (
	"context"
	"net/http"
)

type subscriptionDeleteService interface {
	Delete(ctx context.Context, id int64) error
}

func (h *SubscriptionAdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	deleter, ok := h.services.Subscriptions.(subscriptionDeleteService)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "subscription delete is not configured")
		return
	}

	if err := deleter.Delete(r.Context(), id); err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"id": id,
	})
}
