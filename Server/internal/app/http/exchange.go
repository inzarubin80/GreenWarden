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
	serviceLogin interface {
		Login(ctx context.Context, providerKey string, authorizationCode string, codeVerifier string) (*model.AuthData, error)
	}

	ExchangeHandler struct {
		name    string
		store   *sessions.CookieStore
		service serviceLogin
	}
)

func NewExchangeHandler(store *sessions.CookieStore, name string, service serviceLogin) *ExchangeHandler {
	return &ExchangeHandler{
		name:    name,
		store:   store,
		service: service,
	}
}

func (h *ExchangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Provider string `json:"provider"`
		Code     string `json:"code"`
		CodeVerifier string `json:"code_verifier"`
		State    string `json:"state"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Provider == "" || req.Code == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "provider and code are required")
		return
	}

	authData, err := h.service.Login(r.Context(), req.Provider, req.Code, req.CodeVerifier)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	session.Values[defenitions.Token] = authData.RefreshToken
	session.Values[defenitions.UserID] = int64(authData.UserID)
	if err := session.Save(r, w); err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	type response struct {
		Token  string      `json:"token"`
		UserID model.UserID `json:"user_id"`
	}

	resp := response{
		Token:  authData.AccessToken,
		UserID: authData.UserID,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, jsonData)
}


