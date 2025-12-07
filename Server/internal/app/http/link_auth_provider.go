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
	linkAuthProviderService interface {
		LinkAuthProvider(ctx context.Context, userID model.UserID, providerKey, authorizationCode, codeVerifier string) (*model.UserProfile, error)
	}

	LinkAuthProviderHandler struct {
		name    string
		service linkAuthProviderService
	}
)

func NewLinkAuthProviderHandler(name string, service linkAuthProviderService) *LinkAuthProviderHandler {
	return &LinkAuthProviderHandler{
		name:    name,
		service: service,
	}
}

type linkAuthProviderRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

func (h *LinkAuthProviderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var req linkAuthProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Code == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "code is required")
		return
	}

	profile, err := h.service.LinkAuthProvider(ctx, userID, providerKey, req.Code, req.CodeVerifier)
	if err != nil {
		// Бизнес-ошибки мапим в 400
		if err == service.ErrProviderAlreadyLinkedToAnotherUser {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
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


