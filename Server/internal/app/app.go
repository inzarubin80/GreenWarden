package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"

	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	appHttp "github.com/inzarubin80/Server/internal/app/http"
	middleware "github.com/inzarubin80/Server/internal/app/http/middleware"
	tokenservice "github.com/inzarubin80/Server/internal/app/token_service"
	ws "github.com/inzarubin80/Server/internal/app/ws"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/repository"
	service "github.com/inzarubin80/Server/internal/service"

	//"github.com/rs/cors"
	"golang.org/x/oauth2"
)

const (
	readHeaderTimeoutSeconds = 3
)

type (
	mux interface {
		Handle(pattern string, handler http.Handler)
	}
	server interface {
		ListenAndServe() error
		ListenAndServeTLS(certFile, keyFile string) error
		Close() error
	}

	App struct {
		mux           mux
		server        server
		pokerService  *service.PokerService
		config        config
		hub           *ws.Hub
		oauthConfig   *oauth2.Config
		store         *sessions.CookieStore
		provadersConf authinterface.MapProviderOauthConf
	}
)

// repoAdapter provides stub implementations for methods not needed by the exchange/login flow.
type repoAdapter struct{ *repository.Repository }

// Provide stubs for unused methods to satisfy service.Repository
func (r *repoAdapter) AddPokerUser(ctx context.Context, pokerID model.PokerID, userID model.UserID) error {
	return nil
}

func (r *repoAdapter) GetUserIDsByPokerID(ctx context.Context, pokerID model.PokerID) ([]model.UserID, error) {
	return []model.UserID{}, nil
}

// hubAdapter provides no-op implementations to satisfy the service.Hub interface.
type hubAdapter struct{ *ws.Hub }

func (h *hubAdapter) AddMessage(pokerID model.PokerID, payload any) error { return nil }
func (h *hubAdapter) AddMessageForUser(pokerID model.PokerID, userID model.UserID, payload any) error {
	return nil
}
func (h *hubAdapter) GetActiveUsersID(pokerID model.PokerID) ([]model.UserID, error) {
	return []model.UserID{}, nil
}

func (a *App) ListenAndServe() error {
	go a.hub.Run()

	a.mux.Handle(a.config.path.ping, appHttp.NewPingHandlerHandler(a.config.path.ping))
	a.mux.Handle(a.config.path.session, appHttp.NewGetSessionHandler(a.store, a.config.path.session))
	a.mux.Handle(a.config.path.getProviders, appHttp.NewProvadersHandler(a.provadersConf, a.config.path.getProviders))
	a.mux.Handle(a.config.path.login, appHttp.NewLoginHandler(a.provadersConf, a.config.path.login, a.store))
	a.mux.Handle(a.config.path.exchange, appHttp.NewExchangeHandler(a.store, a.config.path.exchange, a.pokerService))
	fmt.Println("start server")

	return a.server.ListenAndServe()
}

func NewApp(ctx context.Context, config config, dbConn *pgxpool.Pool) (*App, error) {

	var (
		mux   = http.NewServeMux()
		hub   = ws.NewHub()
		store = sessions.NewCookieStore([]byte(config.sectrets.storeSecret))
	)

	// Build repository
	repo := repository.NewPokerRepository(dbConn)

	// Build token services
	accessTokenService := tokenservice.NewtokenService([]byte(config.sectrets.accessTokenSecret), 30*time.Minute, model.Access_Token_Type)
	refreshTokenService := tokenservice.NewtokenService([]byte(config.sectrets.refreshTokenSecret), 30*24*time.Hour, model.Refresh_Token_Type)

	// Build providers user data map from config
	providersMap := make(authinterface.ProvidersUserData)
	for key, prov := range config.provadersConf {
		if prov != nil && prov.ProviderUserData != nil {
			providersMap[key] = prov.ProviderUserData
		}
	}

	// Build service
	pokerService := service.NewPokerService(&repoAdapter{Repository: repo}, &hubAdapter{Hub: hub}, accessTokenService, refreshTokenService, providersMap)

	/*
		// Создаем CORS middleware
		corsMiddleware := cors.New(cors.Options{
			// Явно разрешаем оба домена (без точки в начале)
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://10.0.2.2",
			},
			// Добавляем все необходимые методы
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			// Разрешаем все стандартные заголовки + кастомные
			AllowedHeaders: []string{
				"Origin", "Content-Type", "Accept", "Authorization",
				"X-Requested-With", "X-CSRF-Token", "Custom-Header",
			},
			// Разрешаем куки и авторизацию
			AllowCredentials: true,
			// Опционально: максимальное время кеширования preflight-запросов
			MaxAge: 86400,
		})
	*/

	// Обертываем основной обработчик
	handler := middleware.NewLogMux(mux)

	return &App{
		mux:           mux,
		server:        &http.Server{Addr: config.addr, Handler: handler, ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second},
		pokerService:  pokerService,
		config:        config,
		hub:           hub,
		store:         store,
		provadersConf: config.provadersConf,
	}, nil

}
