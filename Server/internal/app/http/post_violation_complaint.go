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
	postViolationComplaintService interface {
		CreateViolationComplaint(ctx context.Context, userID model.UserID, violationID model.ViolationID, reason, message string) (*model.ViolationComplaint, error)
	}

	PostViolationComplaintHandler struct {
		name    string
		service postViolationComplaintService
	}
)

func NewPostViolationComplaintHandler(name string, service postViolationComplaintService) *PostViolationComplaintHandler {
	return &PostViolationComplaintHandler{name: name, service: service}
}

type postViolationComplaintRequest struct {
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func (h *PostViolationComplaintHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
		return
	}

	var req postViolationComplaintRequest
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

	complaint, err := h.service.CreateViolationComplaint(ctx, userID, model.ViolationID(id), req.Reason, req.Message)
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


