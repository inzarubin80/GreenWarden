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
	serviceUpdateViolationChatMessage interface {
		UpdateViolationChatMessage(ctx context.Context, userID model.UserID, messageID string, text string) (*model.ViolationChatMessage, error)
	}

	UpdateViolationChatMessageHandler struct {
		name    string
		service serviceUpdateViolationChatMessage
	}
)

func NewUpdateViolationChatMessageHandler(name string, service serviceUpdateViolationChatMessage) *UpdateViolationChatMessageHandler {
	return &UpdateViolationChatMessageHandler{name: name, service: service}
}

type updateViolationChatMessageRequest struct {
	Text string `json:"text"`
}

func (h *UpdateViolationChatMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	violationID := r.PathValue("id")
	messageID := r.PathValue("message_id")
	if violationID == "" || messageID == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id and message id are required")
		return
	}

	var req updateViolationChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "text is required")
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	msg, err := h.service.UpdateViolationChatMessage(ctx, userID, messageID, req.Text)
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


