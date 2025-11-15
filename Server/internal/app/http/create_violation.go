package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	createViolationService interface {
		CreateViolation(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64) (*model.Violation, error)
	}

	CreateViolationHandler struct {
		name    string
		store   *sessions.CookieStore
		service createViolationService
	}
)

func NewCreateViolationHandler(store *sessions.CookieStore, name string, service createViolationService) *CreateViolationHandler {
	return &CreateViolationHandler{
		name:    name,
		store:   store,
		service: service,
	}
}

func (h *CreateViolationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Type        string  `json:"type"`
		Description string  `json:"description"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	v, err := h.service.CreateViolation(ctx, userID, model.ViolationType(req.Type), req.Description, req.Lat, req.Lng)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := json.Marshal(v)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, resp)
}


