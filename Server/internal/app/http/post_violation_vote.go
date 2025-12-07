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
	postViolationVoteService interface {
		SetViolationVote(ctx context.Context, userID model.UserID, violationID model.ViolationID, value string) (*model.ViolationVotes, error)
	}

	PostViolationVoteHandler struct {
		name    string
		service postViolationVoteService
	}
)

func NewPostViolationVoteHandler(name string, service postViolationVoteService) *PostViolationVoteHandler {
	return &PostViolationVoteHandler{name: name, service: service}
}

type postViolationVoteRequest struct {
	Value string `json:"value"`
}

func (h *PostViolationVoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
		return
	}

	var req postViolationVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Value == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "value is required")
		return
	}

	// value: "like" | "dislike" | "none"
	if req.Value != "like" && req.Value != "dislike" && req.Value != "none" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid value")
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	votes, err := h.service.SetViolationVote(ctx, userID, model.ViolationID(id), req.Value)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	b, err := json.Marshal(votes)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, b)
}


