package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/service"
)

type OAuthCallbackHandler struct {
	name              string
	provadersConf     authinterface.MapProviderOauthConf
	store             *sessions.CookieStore
	loginStateStore   map[string]StateData
	loginStateStoreMu *sync.Mutex
	service           serviceLogin
	linkService       linkAuthProviderService
}

func NewOAuthCallbackHandler(
	provadersConf authinterface.MapProviderOauthConf,
	name string,
	store *sessions.CookieStore,
	loginStateStore map[string]StateData,
	loginStateStoreMu *sync.Mutex,
	service serviceLogin,
	linkService linkAuthProviderService,
) *OAuthCallbackHandler {
	return &OAuthCallbackHandler{
		name:              name,
		provadersConf:     provadersConf,
		store:             store,
		loginStateStore:   loginStateStore,
		loginStateStoreMu: loginStateStoreMu,
		service:           service,
		linkService:       linkService,
	}
}

func (h *OAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из query string
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	provider := r.URL.Query().Get("provider")
	errorParam := r.URL.Query().Get("error")
	errorDescription := r.URL.Query().Get("error_description")

	// Определяем action из state (будет получен позже) или используем "login" по умолчанию
	// Временно используем пустую строку, action будет определен после получения stateInfo
	action := ""

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
	action = stateInfo.Action
	if action == "" {
		action = "login" // значение по умолчанию для обратной совместимости
	}
	delete(h.loginStateStore, state)
	h.loginStateStoreMu.Unlock()

	// Различаем логин и привязку провайдера
	if action == "link" {
		// Привязка провайдера - проверяем наличие активной сессии
		session, err := h.store.Get(r, defenitions.SessionAuthenticationName)
		if err != nil {
			mobileRedirect := fmt.Sprintf("warden://auth/callback?action=link&provider=%s&error=unauthorized&error_description=%s",
				provider, url.QueryEscape("session_not_found"))
			http.Redirect(w, r, mobileRedirect, http.StatusFound)
			return
		}

		userIDValue := session.Values[defenitions.UserID]
		if userIDValue == nil {
			mobileRedirect := fmt.Sprintf("warden://auth/callback?action=link&provider=%s&error=unauthorized&error_description=%s",
				provider, url.QueryEscape("user_not_authenticated"))
			http.Redirect(w, r, mobileRedirect, http.StatusFound)
			return
		}

		userID, ok := userIDValue.(int64)
		if !ok || userID == 0 {
			mobileRedirect := fmt.Sprintf("warden://auth/callback?action=link&provider=%s&error=unauthorized&error_description=%s",
				provider, url.QueryEscape("invalid_user_id"))
			http.Redirect(w, r, mobileRedirect, http.StatusFound)
			return
		}

		// Вызываем LinkAuthProvider
		_, err = h.linkService.LinkAuthProvider(r.Context(), model.UserID(userID), provider, code, codeVerifier)
		if err != nil {
			// Маппинг ошибок привязки
			errorCode := "link_failed"
			errorMsg := err.Error()
			
			// Проверяем специфичные ошибки
			if errors.Is(err, service.ErrProviderAlreadyLinkedToAnotherUser) {
				errorCode = "provider_already_linked"
			} else if errorMsg == "provider not found" {
				errorCode = "provider_not_found"
			} else if errorMsg == "exchange_failed" || errorMsg == "token exchange failed" {
				errorCode = "exchange_failed"
			}

			mobileRedirect := fmt.Sprintf("warden://auth/callback?action=link&provider=%s&error=%s&error_description=%s",
				provider, errorCode, url.QueryEscape(errorMsg))
			http.Redirect(w, r, mobileRedirect, http.StatusFound)
			return
		}

		// Успешная привязка
		mobileRedirect := fmt.Sprintf("warden://auth/callback?action=link&provider=%s&success=true",
			provider)
		http.Redirect(w, r, mobileRedirect, http.StatusFound)
		return
	}

	// Обычный логин (action == "login" или отсутствует)
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
