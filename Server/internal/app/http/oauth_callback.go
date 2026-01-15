package http

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/model"
)

type OAuthCallbackHandler struct {
	name              string
	provadersConf     authinterface.MapProviderOauthConf
	store             *sessions.CookieStore
	loginStateStore   map[string]StateData
	loginStateStoreMu *sync.Mutex
	service           serviceLogin
}

func NewOAuthCallbackHandler(
	provadersConf authinterface.MapProviderOauthConf,
	name string,
	store *sessions.CookieStore,
	loginStateStore map[string]StateData,
	loginStateStoreMu *sync.Mutex,
	service serviceLogin,
) *OAuthCallbackHandler {
	return &OAuthCallbackHandler{
		name:              name,
		provadersConf:     provadersConf,
		store:             store,
		loginStateStore:   loginStateStore,
		loginStateStoreMu: loginStateStoreMu,
		service:           service,
	}
}

func (h *OAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из query string
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	provider := r.URL.Query().Get("provider")
	errorParam := r.URL.Query().Get("error")
	errorDescription := r.URL.Query().Get("error_description")

	// Если есть ошибка от провайдера
	if errorParam != "" {
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=%s", provider, url.QueryEscape(errorParam))
		if errorDescription != "" {
			mobileRedirect += fmt.Sprintf("&error_description=%s", url.QueryEscape(errorDescription))
		}
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	if code == "" || state == "" {
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=invalid_request&error_description=%s",
			provider, url.QueryEscape("missing_code_or_state"))
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	// Получаем сохраненные данные по state
	h.loginStateStoreMu.Lock()
	stateInfo, ok := h.loginStateStore[state]
	if !ok {
		h.loginStateStoreMu.Unlock()
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=invalid_state&error_description=%s",
			provider, url.QueryEscape("state_not_found"))
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	// Проверяем expiry
	if time.Now().After(stateInfo.Expiry) {
		delete(h.loginStateStore, state)
		h.loginStateStoreMu.Unlock()
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=expired_state&error_description=%s",
			provider, url.QueryEscape("state_expired"))
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	// Проверяем provider
	if stateInfo.Provider != provider {
		delete(h.loginStateStore, state)
		h.loginStateStoreMu.Unlock()
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=invalid_provider&error_description=%s",
			provider, url.QueryEscape("provider_mismatch"))
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	codeVerifier := stateInfo.CodeVerifier
	delete(h.loginStateStore, state)
	h.loginStateStoreMu.Unlock()

	// Обмениваем код на токены
	var authData *model.AuthData
	authData, err := h.service.Login(r.Context(), provider, code, codeVerifier)
	if err != nil {
		mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&error=exchange_failed&error_description=%s",
			provider, url.QueryEscape(err.Error()))
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	// Сохраняем сессию
	session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
	if err == nil {
		session.Values[defenitions.Token] = authData.RefreshToken
		session.Values[defenitions.UserID] = int64(authData.UserID)
		session.Save(r, w)
	}

	// Перенаправляем на мобильное приложение с токенами
	mobileRedirect := fmt.Sprintf("warden://auth/callback?provider=%s&access_token=%s&user_id=%d",
		provider, url.QueryEscape(authData.AccessToken), authData.UserID)
	http.Redirect(w, r, mobileRedirect, http.StatusFound)
}
