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
	updateProfileService interface {
		UpdateUserProfile(ctx context.Context, userID model.UserID, name *string, boostyURL *string) (*model.UserProfile, error)
	}

	UpdateProfileHandler struct {
		name    string
		service updateProfileService
	}
)

func NewUpdateProfileHandler(name string, service updateProfileService) *UpdateProfileHandler {
	return &UpdateProfileHandler{
		name:    name,
		service: service,
	}
}

func (h *UpdateProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	// Нам важно различать:
	// - поле отсутствует (не менять),
	// - поле есть и равно null (очистить),
	// - поле есть и строка (обновить).
	// Поэтому сначала декодируем в raw map.
	var raw map[string]*json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	var (
		namePtr      *string
		boostyURLPtr *string
	)

	// display_name
	if v, ok := raw["display_name"]; ok {
		if v == nil || string(*v) == "null" {
			empty := ""
			namePtr = &empty
		} else {
			var s string
			if err := json.Unmarshal(*v, &s); err == nil {
				namePtr = &s
			}
		}
	}

	// boosty_url
	if v, ok := raw["boosty_url"]; ok {
		if v == nil || string(*v) == "null" {
			empty := ""
			boostyURLPtr = &empty
		} else {
			var s string
			if err := json.Unmarshal(*v, &s); err == nil {
				boostyURLPtr = &s
			}
		}
	}

	profile, err := h.service.UpdateUserProfile(ctx, userID, namePtr, boostyURLPtr)
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
