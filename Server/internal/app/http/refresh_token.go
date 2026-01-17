package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	serviceRefreshToken interface {
		RefreshToken(ctx context.Context, refreshToken string) (*model.AuthData, error)
	}

	RefreshTokenHandler struct {
		name    string
		service serviceRefreshToken
		store   *sessions.CookieStore
	}
)

func NewRefreshTokenHandler(service serviceRefreshToken, name string, store *sessions.CookieStore) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		name:    name,
		service: service,
		store:   store,
	}
}

func (h *RefreshTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var tokenString string
	var useCookie bool

	// 1. Попытаться получить refresh_token из JSON body
	bodyBytes, err := io.ReadAll(r.Body)
	if err == nil && len(bodyBytes) > 0 {
		// Восстанавливаем body для возможного повторного чтения
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.Unmarshal(bodyBytes, &body); err == nil && body.RefreshToken != "" {
			tokenString = body.RefreshToken
			useCookie = false
		}
	}

	// 2. Fallback на cookie/сессию (для web-клиентов)
	if tokenString == "" {
		session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
		if err != nil {
			http.Error(w, "Unauthorized not session", http.StatusUnauthorized)
			return
		}

		var ok bool
		tokenString, ok = session.Values[defenitions.Token].(string)
		if !ok {
			http.Error(w, "Unauthorized not Token", http.StatusUnauthorized)
			return
		}
		useCookie = true
	}

	authData, err := h.service.RefreshToken(ctx, tokenString)
	if err != nil {
		http.Error(w, "Unauthorized not session", http.StatusUnauthorized)
		return
	}

	// Обновляем cookie только если использовали cookie для получения токена
	if useCookie {
		session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
		if err == nil {
			session.Values[defenitions.Token] = string(authData.RefreshToken)
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	type response struct {
		Token        string       `json:"token"`
		RefreshToken string       `json:"refresh_token"`
		UserID       model.UserID `json:"user_id"`
	}

	resp := response{
		Token:        authData.AccessToken,
		RefreshToken: authData.RefreshToken,
		UserID:       authData.UserID,
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uhttp.SendSuccessfulResponse(w, jsonData)
}
