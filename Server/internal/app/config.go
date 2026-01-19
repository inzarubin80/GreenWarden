package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	providerUserData "github.com/inzarubin80/Server/internal/app/clients/provider_user_data"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/icons"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type (
	Options struct {
		Addr string
	}
	path struct {
		index, getPoker, createPoker, createTask,
		getTasks, getTask, updateTask, deleteTask,
		getComents, addComent, setVotingTask,
		getVotingControlState, ws, login, exchange, createViolation, listViolations, getViolation, getViolationRequest, closeViolationRequest, session, refreshToken, logOut, getProviders,
		violationChatWS, getViolationChat, postViolationChatMessage, updateViolationChatMessage, deleteViolationChatMessage,
		getViolationVote, postViolationVote, postViolationComplaint,
		postViolationRequestVote, postViolationRequestComplaint,
		ping, vote, getUserEstimates, setVotingControlState, setUserName, getUser, setUserSettings, getLastSession, deletePoker,
		getProfile, updateProfile, uploadProfileAvatar,
		linkAuthProvider, unlinkAuthProvider, oauthCallback string
		// Public share pages (for OG tags)
		violationSharePage string
	}

	sectrets struct {
		storeSecret           string
		accessTokenSecret     string
		refreshTokenSecret    string
		mobileAppSecret       string
		yosAccessKey          string
		yosSecretKey          string
		yosBucket             string
		yosEndpoint           string
		yosCdnBaseURL         string
		maxPhotosPerViolation int
	}

	config struct {
		addr          string
		path          path
		sectrets      sectrets
		provadersConf authinterface.MapProviderOauthConf
		// TLS debug settings
		tlsEnabled  bool
		tlsCertFile string
		tlsKeyFile  string
		// CORS settings
		corsAllowedOrigins []string
	}
)

func NewConfig(opts Options) config {
	provaders := make(authinterface.MapProviderOauthConf)
	// API_ROOT - адрес сервера для OAuth redirect URI
	apiRoot := os.Getenv("API_ROOT")
	if apiRoot == "" {
		apiRoot = "https://api.green-warden.ru" // fallback для продакшена
	}
	provaders["yandex"] = &authinterface.ProviderOauthConf{
		Oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_YANDEX"),
			ClientSecret: os.Getenv("CLIENT_SECRET_YANDEX"),
			RedirectURL:  apiRoot + "/api/auth/callback?provider=yandex",
			Scopes:       []string{"login:info"},
			Endpoint:     yandex.Endpoint,
		},
		UrlUserData: "https://login.yandex.ru/info?format=json",
		IconSVG:     icons.GetProviderIcon("yandex"),
		DisplayName: "Яндекс",
		ProviderUserData: providerUserData.NewProviderUserData("https://login.yandex.ru/info?format=json", &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_YANDEX"),
			ClientSecret: os.Getenv("CLIENT_SECRET_YANDEX"),
			RedirectURL:  apiRoot + "/api/auth/callback?provider=yandex",
			Scopes:       []string{"login:info"},
			Endpoint:     yandex.Endpoint,
		}, "yandex"),
	}

	// Добавим Google провайдер для демонстрации
	provaders["google"] = &authinterface.ProviderOauthConf{
		Oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_GOOGLE"),
			ClientSecret: os.Getenv("CLIENT_SECRET_GOOGLE"),
			RedirectURL:  apiRoot + "/api/auth/callback?provider=google",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
		UrlUserData: "https://www.googleapis.com/oauth2/v2/userinfo",
		IconSVG:     icons.GetProviderIcon("google"),
		DisplayName: "Google",
		ProviderUserData: providerUserData.NewProviderUserData("https://www.googleapis.com/oauth2/v2/userinfo", &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_GOOGLE"),
			ClientSecret: os.Getenv("CLIENT_SECRET_GOOGLE"),
			RedirectURL:  apiRoot + "/api/auth/callback?provider=google",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		}, "google"),
	}

	// Парсим CORS allowed origins из переменной окружения
	corsOrigins := []string{} // Начинаем с пустого массива
	if corsEnv := os.Getenv("CORS_ALLOWED_ORIGINS"); corsEnv != "" {
		// Разделяем по запятой и очищаем пробелы
		origins := strings.Split(corsEnv, ",")
		corsOrigins = make([]string, 0, len(origins))
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				corsOrigins = append(corsOrigins, origin)
			}
		}
	} else {
		// Если переменная не установлена, используем дефолтные значения
		isProduction := os.Getenv("ENV") == "production"
		if isProduction {
			// Для продакшена: добавляем продакшен origin
			corsOrigins = []string{
				"https://green-warden.ru",
			}
		} else {
			// Для разработки: добавляем дефолтные origins
			corsOrigins = []string{
				"http://localhost:3000",
				"http://10.0.2.2", // Android emulator
			}
		}
	}

	config := config{
		addr: opts.Addr,
		path: path{
			index:        "",
			ping:         "GET /api/ping",
			createPoker:  "POST	/api/poker",
			getProviders: "GET /api/providers",

			login:                 "POST	/api/user/login",
			exchange:              "POST	/api/user/exchange",
			createViolation:       "POST	/api/violations",
			listViolations:        "GET /api/violations",
			getViolation:          "GET /api/violations/{id}",
			getViolationRequest:   "GET /api/violations/{violation_id}/requests/{request_id}",
			closeViolationRequest: "POST /api/violations/{id}/close-request",
			setUserName:           "POST	/api/user/name",
			setUserSettings:       "POST	/api/user/settings",

			getUser: "GET	/api/user",

			getProfile:          "GET /api/profile",
			updateProfile:       "PATCH /api/profile",
			uploadProfileAvatar: "POST /api/profile/avatar",

			linkAuthProvider:   "POST /api/profile/auth-providers/{provider}/link",
			unlinkAuthProvider: "POST /api/profile/auth-providers/{provider}/unlink",

			refreshToken:  "POST	/api/user/refresh",
			session:       "GET		/api/user/session",
			logOut:        "GET		/api/user/logout",
			oauthCallback: "GET /api/auth/callback",

			violationChatWS:            "GET /api/ws/violation-chat",
			getViolationChat:           "GET /api/violations/{id}/chat",
			postViolationChatMessage:   "POST /api/violations/{id}/chat/messages",
			updateViolationChatMessage: "PATCH /api/violations/{id}/chat/messages/{message_id}",
			deleteViolationChatMessage: "DELETE /api/violations/{id}/chat/messages/{message_id}",

			getViolationVote:              "GET /api/violations/{id}/vote",
			postViolationVote:             "POST /api/violations/{id}/vote",
			postViolationComplaint:        "POST /api/violations/{id}/complaints",
			postViolationRequestVote:      "POST /api/violation-requests/{id}/vote",
			postViolationRequestComplaint: "POST /api/violation-requests/{id}/complaints",

			// Public share page for social previews (VK/OG)
			violationSharePage: "GET /violations/{id}",

			getLastSession: fmt.Sprintf("GET	/api/sessions/{%s}/{%s}", defenitions.Page, defenitions.PageSize),
		},

		sectrets: sectrets{
			storeSecret:        os.Getenv("STORE_SECRET"),
			accessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
			refreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
			mobileAppSecret:    os.Getenv("MOBILE_APP_SECRET"),

			yosAccessKey:  os.Getenv("YOS_ACCESS_KEY"),
			yosSecretKey:  os.Getenv("YOS_SECRET_KEY"),
			yosBucket:     os.Getenv("YOS_BUCKET"),
			yosEndpoint:   os.Getenv("YOS_ENDPOINT"),
			yosCdnBaseURL: os.Getenv("YOS_CDN_BASE_URL"),

			maxPhotosPerViolation: func() int {
				if v := os.Getenv("MAX_PHOTOS_PER_VIOLATION"); v != "" {
					if n, err := strconv.Atoi(v); err == nil {
						return n
					}
				}
				return 5
			}(),
		},

		provadersConf:      provaders,
		tlsEnabled:         true,
		tlsCertFile:        os.Getenv("TLS_CERT_FILE"),
		tlsKeyFile:         os.Getenv("TLS_KEY_FILE"),
		corsAllowedOrigins: corsOrigins,
	}

	return config
}
