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
	postViolationRequestComplaintService interface {
		CreateViolationRequestComplaint(ctx context.Context, userID model.UserID, requestID string, reason, message string) (*model.ViolationComplaint, error)
	}

	PostViolationRequestComplaintHandler struct {
		name    string
		service postViolationRequestComplaintService
	}
)

func NewPostViolationRequestComplaintHandler(name string, service postViolationRequestComplaintService) *PostViolationRequestComplaintHandler {
	return &PostViolationRequestComplaintHandler{name: name, service: service}
}

type postViolationRequestComplaintRequest struct {
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func (h *PostViolationRequestComplaintHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation request id is required")
		return
	}

	var req postViolationRequestComplaintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	complaint, err := h.service.CreateViolationRequestComplaint(ctx, userID, id, req.Reason, req.Message)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	b, err := json.Marshal(complaint)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	uhttp.SendSuccessfulResponse(w, b)
}


