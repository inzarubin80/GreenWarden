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
	getProfileService interface {
		GetUserProfile(ctx context.Context, userID model.UserID) (*model.UserProfile, error)
	}

	GetProfileHandler struct {
		name    string
		service getProfileService
	}
)

func NewGetProfileHandler(name string, service getProfileService) *GetProfileHandler {
	return &GetProfileHandler{
		name:    name,
		service: service,
	}
}

func (h *GetProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	profile, err := h.service.GetUserProfile(ctx, userID)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonData, err := json.Marshal(profile)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, jsonData)
}


