package http

import (
	"context"
	"net/http"

	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	serviceDeleteViolationChatMessage interface {
		DeleteViolationChatMessage(ctx context.Context, userID model.UserID, messageID string) error
	}

	DeleteViolationChatMessageHandler struct {
		name    string
		service serviceDeleteViolationChatMessage
	}
)

func NewDeleteViolationChatMessageHandler(name string, service serviceDeleteViolationChatMessage) *DeleteViolationChatMessageHandler {
	return &DeleteViolationChatMessageHandler{name: name, service: service}
}

func (h *DeleteViolationChatMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	violationID := r.PathValue("id")
	messageID := r.PathValue("message_id")
	if violationID == "" || messageID == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id and message id are required")
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.DeleteViolationChatMessage(ctx, userID, messageID); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


