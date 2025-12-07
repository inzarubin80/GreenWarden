package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	servicePostViolationChatMessage interface {
		PostViolationChatMessage(ctx context.Context, userID model.UserID, violationID model.ViolationID, text string, isSystem bool) (*model.ViolationChatMessage, error)
	}

	PostViolationChatMessageHandler struct {
		name    string
		service servicePostViolationChatMessage
	}
)

func NewPostViolationChatMessageHandler(name string, service servicePostViolationChatMessage) *PostViolationChatMessageHandler {
	return &PostViolationChatMessageHandler{name: name, service: service}
}

type postViolationChatMessageRequest struct {
	Text string `json:"text"`
}

func (h *PostViolationChatMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
		return
	}

	var req postViolationChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "text is required")
		return
	}

	// Извлекаем user_id из контекста, проставленный AuthMiddleware.
	ctx := r.Context()
	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	msg, err := h.service.PostViolationChatMessage(ctx, userID, model.ViolationID(id), req.Text, false)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	b, err := json.Marshal(msg)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	uhttp.SendSuccessfulResponse(w, b)
}


