package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	getViolationVoteService interface {
		GetViolationVotes(ctx context.Context, userID model.UserID, violationID model.ViolationID) (*model.ViolationVotes, error)
	}

	GetViolationVoteHandler struct {
		name    string
		service getViolationVoteService
	}
)

func NewGetViolationVoteHandler(name string, service getViolationVoteService) *GetViolationVoteHandler {
	return &GetViolationVoteHandler{name: name, service: service}
}

func (h *GetViolationVoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
		return
	}

	// Для упрощения: не извлекаем user_id, всегда возвращаем user_vote = "".
	votes, err := h.service.GetViolationVotes(r.Context(), 0, model.ViolationID(id))
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


