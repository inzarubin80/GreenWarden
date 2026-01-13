package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"net/url"

	"github.com/gorilla/sessions"
	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	"github.com/inzarubin80/Server/internal/app/uhttp"
)

type LoginHandler struct {
	name              string
	provadersConf     authinterface.MapProviderOauthConf
	store             *sessions.CookieStore
	loginStateStore   map[string]time.Time
	loginStateStoreMu sync.Mutex
}

func NewLoginHandler(provadersConf authinterface.MapProviderOauthConf, name string, store *sessions.CookieStore) *LoginHandler {
	return &LoginHandler{
		name:              name,
		provadersConf:     provadersConf,
		store:             store,
		loginStateStore:   make(map[string]time.Time),
		loginStateStoreMu: sync.Mutex{},
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Provider      string `json:"provider"`
		CodeChallenge string `json:"code_challenge"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Provider == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "provider required")
		return
	}

	cfg, ok := h.provadersConf[req.Provider]
	if !ok || cfg == nil || cfg.Oauth2Config == nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "unknown provider")
		return
	}

	state := randomURLSafe(24)

	// handle PKCE: prefer client-provided code_challenge; if not and pkce requested, generate verifier
	var challenge string
	// require client-provided code_challenge (mobile generates verifier)
	if req.CodeChallenge == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "code_challenge required from client")
		return
	}
	challenge = req.CodeChallenge

	// save state server-side (one-time, TTL 15 minutes) so we can validate it at exchange
	h.loginStateStoreMu.Lock()
	h.loginStateStore[state] = time.Now().Add(15 * time.Minute)
	h.loginStateStoreMu.Unlock()

	// Build auth_url explicitly to ensure redirect_uri matches mobile registration.
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI_MOBILE")
	if redirectURI == "" {
		if cfg.Oauth2Config != nil && cfg.Oauth2Config.RedirectURL != "" {
			redirectURI = cfg.Oauth2Config.RedirectURL
		} else {
			redirectURI = fmt.Sprintf("warden://auth/callback?provider=%s", req.Provider)
		}
	}

	scope := "login:info"
	if cfg.Oauth2Config != nil && len(cfg.Oauth2Config.Scopes) > 0 {
		// Join all scopes with space separator
		scope = ""
		for i, s := range cfg.Oauth2Config.Scopes {
			if i > 0 {
				scope += " "
			}
			scope += s
		}
	}

	// build OAuth authorize URL using provider's endpoint
	base := cfg.Oauth2Config.Endpoint.AuthURL
	q := make(url.Values)
	q.Set("client_id", cfg.Oauth2Config.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", scope)
	q.Set("state", state)
	if challenge != "" {
		q.Set("code_challenge", challenge)
		q.Set("code_challenge_method", "S256")
	}

	authURL := base + "?" + q.Encode()

	resp := map[string]string{
		"auth_url": authURL,
		"state":    state,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	uhttp.SendSuccessfulResponse(w, b)
}

func randomURLSafe(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	s := base64.RawURLEncoding.EncodeToString(b)
	if len(s) > n {
		return s[:n]
	}
	return s
}
