package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	serviceGetViolationChat interface {
		GetViolationChat(ctx context.Context, violationID model.ViolationID, page, pageSize int) (*model.PaginatedViolationChatMessages, error)
	}

	GetViolationChatHandler struct {
		name    string
		service serviceGetViolationChat
	}
)

func NewGetViolationChatHandler(name string, service serviceGetViolationChat) *GetViolationChatHandler {
	return &GetViolationChatHandler{name: name, service: service}
}

func (h *GetViolationChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
		return
	}

	q := r.URL.Query()
	page := 1
	pageSize := 50
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page = n
		}
	}
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pageSize = n
		}
	}

	res, err := h.service.GetViolationChat(r.Context(), model.ViolationID(id), page, pageSize)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	uhttp.SendSuccessfulResponse(w, b)
}


