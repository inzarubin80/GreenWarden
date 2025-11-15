package app

import (
	"fmt"
	"os"

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
		getVotingControlState, ws, login, exchange, createViolation, session, refreshToken, logOut, getProviders,
		ping, vote, getUserEstimates, setVotingControlState, setUserName, getUser, setUserSettings, getLastSession, deletePoker string
	}

	sectrets struct {
		storeSecret        string
		accessTokenSecret  string
		refreshTokenSecret string
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
	}
)

func NewConfig(opts Options) config {
	provaders := make(authinterface.MapProviderOauthConf)
	provaders["yandex"] = &authinterface.ProviderOauthConf{
		Oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_YANDEX"),
			ClientSecret: os.Getenv("CLIENT_SECRET_YANDEX"),
			RedirectURL:  "warden://auth/callback?provider=yandex",
			Scopes:       []string{"login:info"},
			Endpoint:     yandex.Endpoint,
		},
		UrlUserData: "https://login.yandex.ru/info?format=json",
		IconSVG:     icons.GetProviderIcon("yandex"),
		DisplayName: "Яндекс",
		ProviderUserData: providerUserData.NewProviderUserData("https://login.yandex.ru/info?format=json", &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_YANDEX"),
			ClientSecret: os.Getenv("CLIENT_SECRET_YANDEX"),
			RedirectURL:  "warden://auth/callback?provider=yandex",
			Scopes:       []string{"login:info"},
			Endpoint:     yandex.Endpoint,
		}, "yandex"),
	}

	// Добавим Google провайдер для демонстрации
	provaders["google"] = &authinterface.ProviderOauthConf{
		Oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID_GOOGLE"),
			ClientSecret: os.Getenv("CLIENT_SECRET_GOOGLE"),
			RedirectURL:  os.Getenv("APP_ROOT") + "/auth/callback?provider=google",
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
			RedirectURL:  os.Getenv("APP_ROOT") + "/auth/callback?provider=google",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		}, "google"),
	}

	config := config{
		addr: opts.Addr,
		path: path{
			index:        "",
			ping:         "GET /api/ping",
			createPoker:  "POST	/api/poker",
			getProviders: "GET /api/providers",

			login:           "POST	/api/user/login",
			exchange:        "POST	/api/user/exchange",
			createViolation: "POST	/api/violations",
			setUserName:     "POST	/api/user/name",
			setUserSettings: "POST	/api/user/settings",

			getUser: "GET	/api/user",

			refreshToken: "POST	/api/user/refresh",
			session:      "GET		/api/user/session",
			logOut:       "GET		/api/user/logout",

			getLastSession: fmt.Sprintf("GET	/api/sessions/{%s}/{%s}", defenitions.Page, defenitions.PageSize),
		},

		sectrets: sectrets{
			storeSecret:        os.Getenv("STORE_SECRET"),
			accessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
			refreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
		},

		provadersConf: provaders,
		tlsEnabled:    true,
		tlsCertFile:   os.Getenv("TLS_CERT_FILE"),
		tlsKeyFile:    os.Getenv("TLS_KEY_FILE"),
	}

	return config
}
