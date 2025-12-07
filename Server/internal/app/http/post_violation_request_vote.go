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
	postViolationRequestVoteService interface {
		SetViolationRequestVote(ctx context.Context, userID model.UserID, requestID string, value string) (*model.ViolationVotes, error)
	}

	PostViolationRequestVoteHandler struct {
		name    string
		service postViolationRequestVoteService
	}
)

func NewPostViolationRequestVoteHandler(name string, service postViolationRequestVoteService) *PostViolationRequestVoteHandler {
	return &PostViolationRequestVoteHandler{name: name, service: service}
}

type postViolationRequestVoteRequest struct {
	Value string `json:"value"`
}

type postViolationRequestVoteResponse struct {
	ViolationRequestID string `json:"violation_request_id"`
	Likes              int64  `json:"likes"`
	Dislikes           int64  `json:"dislikes"`
	UserVote           string `json:"user_vote"`
}

func (h *PostViolationRequestVoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation request id is required")
		return
	}

	var req postViolationRequestVoteRequest
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

	votes, err := h.service.SetViolationRequestVote(ctx, userID, id, req.Value)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := postViolationRequestVoteResponse{
		ViolationRequestID: id,
		Likes:              votes.Likes,
		Dislikes:           votes.Dislikes,
		UserVote:           votes.UserVote,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, b)
}


