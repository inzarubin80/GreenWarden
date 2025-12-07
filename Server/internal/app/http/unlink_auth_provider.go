package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/service"
)

type (
	unlinkAuthProviderService interface {
		UnlinkAuthProvider(ctx context.Context, userID model.UserID, providerKey string) (*model.UserProfile, error)
	}

	UnlinkAuthProviderHandler struct {
		name    string
		service unlinkAuthProviderService
	}
)

func NewUnlinkAuthProviderHandler(name string, service unlinkAuthProviderService) *UnlinkAuthProviderHandler {
	return &UnlinkAuthProviderHandler{
		name:    name,
		service: service,
	}
}

func (h *UnlinkAuthProviderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok || userID == 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	providerKey := r.PathValue("provider")
	if providerKey == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "provider is required")
		return
	}

	profile, err := h.service.UnlinkAuthProvider(ctx, userID, providerKey)
	if err != nil {
		switch err {
		case service.ErrLastAuthProvider:
			uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		case service.ErrProviderNotLinked:
			uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		default:
			uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	jsonData, err := json.Marshal(profile)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, jsonData)
}


