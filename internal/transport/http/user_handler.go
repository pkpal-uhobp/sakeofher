package httptransport

import (
	"context"
	"net/http"
	"strconv"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type UserHandler struct {
	services *service.Services
}

func NewUserHandler(services *service.Services) *UserHandler {
	return &UserHandler{services: services}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	input := domain.UserListInput{
		Query:  r.URL.Query().Get("query"),
		Status: domain.UserStatus(r.URL.Query().Get("status")),
		Limit:  limit,
		Offset: offset,
	}

	out, err := h.services.Users.List(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, out)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	user, err := h.services.Users.GetByID(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	var input domain.UpdateUserInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.Users.Update(r.Context(), id, input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Block(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.services.Users.Block)
}

func (h *UserHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.services.Users.Unblock)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.services.Users.MarkDeleted)
}

func (h *UserHandler) changeStatus(
	w http.ResponseWriter,
	r *http.Request,
	fn func(rctx context.Context, id int64) (*domain.User, error),
) {
	id, err := pathInt64(r, "id")
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	user, err := fn(r.Context(), id)
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}
