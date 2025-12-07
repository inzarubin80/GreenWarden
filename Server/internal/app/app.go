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
	"github.com/inzarubin80/Server/internal/storage/objectstorage"

	"github.com/rs/cors"
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
		uploader      *objectstorage.Uploader
	}
)

// hubAdapter provides no-op implementations to satisfy the service.Hub interface.
type hubAdapter struct{ *ws.Hub }

func (h *hubAdapter) AddMessage(pokerID model.PokerID, payload any) error { return nil }
func (h *hubAdapter) AddMessageForUser(pokerID model.PokerID, userID model.UserID, payload any) error {
	return nil
}
func (h *hubAdapter) GetActiveUsersID(pokerID model.PokerID) ([]model.UserID, error) {
	return []model.UserID{}, nil
}
func (h *hubAdapter) BroadcastViolationMessage(violationID model.ViolationID, payload any) error {
	if h.Hub == nil {
		return nil
	}
	return h.Hub.BroadcastViolationMessage(violationID, payload)
}
func (h *hubAdapter) SendViolationMessageToUser(violationID model.ViolationID, userID model.UserID, payload any) error {
	if h.Hub == nil {
		return nil
	}
	return h.Hub.SendViolationMessageToUser(violationID, userID, payload)
}

func (a *App) ListenAndServe() error {
	go a.hub.Run()

	a.mux.Handle(a.config.path.ping, appHttp.NewPingHandlerHandler(a.config.path.ping))
	a.mux.Handle(a.config.path.session, appHttp.NewGetSessionHandler(a.store, a.config.path.session))
	a.mux.Handle(a.config.path.getProviders, appHttp.NewProvadersHandler(a.provadersConf, a.config.path.getProviders))
	a.mux.Handle(a.config.path.login, appHttp.NewLoginHandler(a.provadersConf, a.config.path.login, a.store))
	a.mux.Handle(a.config.path.exchange, appHttp.NewExchangeHandler(a.store, a.config.path.exchange, a.pokerService))
	a.mux.Handle(a.config.path.refreshToken, appHttp.NewRefreshTokenHandler(a.pokerService, a.config.path.refreshToken, a.store))
	a.mux.Handle(a.config.path.listViolations, appHttp.NewListViolationsHandler(a.config.path.listViolations, a.pokerService))
	a.mux.Handle(a.config.path.getViolation, middleware.NewAuthMiddleware(appHttp.NewGetViolationHandler(a.config.path.getViolation, a.pokerService, a.uploader), a.store, a.pokerService))
	a.mux.Handle(a.config.path.getViolationRequest, appHttp.NewGetViolationRequestHandler(a.config.path.getViolationRequest, a.pokerService, a.uploader))
	a.mux.Handle(a.config.path.createViolation, middleware.NewAuthMiddleware(appHttp.NewCreateViolationHandler(a.store, a.config.path.createViolation, a.pokerService, a.uploader, a.config.sectrets.maxPhotosPerViolation), a.store, a.pokerService))
	a.mux.Handle(a.config.path.closeViolationRequest, middleware.NewAuthMiddleware(appHttp.NewCloseViolationRequestHandler(a.store, a.config.path.closeViolationRequest, a.pokerService, a.uploader, a.config.sectrets.maxPhotosPerViolation), a.store, a.pokerService))
	a.mux.Handle(a.config.path.violationChatWS, appHttp.NewViolationChatWSHandler(a.config.path.violationChatWS, a.pokerService, a.hub))
	a.mux.Handle(a.config.path.getViolationChat, appHttp.NewGetViolationChatHandler(a.config.path.getViolationChat, a.pokerService))
	a.mux.Handle(a.config.path.postViolationChatMessage, middleware.NewAuthMiddleware(appHttp.NewPostViolationChatMessageHandler(a.config.path.postViolationChatMessage, a.pokerService), a.store, a.pokerService))
	a.mux.Handle(a.config.path.updateViolationChatMessage, middleware.NewAuthMiddleware(appHttp.NewUpdateViolationChatMessageHandler(a.config.path.updateViolationChatMessage, a.pokerService), a.store, a.pokerService))
	a.mux.Handle(a.config.path.deleteViolationChatMessage, middleware.NewAuthMiddleware(appHttp.NewDeleteViolationChatMessageHandler(a.config.path.deleteViolationChatMessage, a.pokerService), a.store, a.pokerService))
	a.mux.Handle(a.config.path.getViolationVote, appHttp.NewGetViolationVoteHandler(a.config.path.getViolationVote, a.pokerService))
	a.mux.Handle(a.config.path.postViolationVote, middleware.NewAuthMiddleware(appHttp.NewPostViolationVoteHandler(a.config.path.postViolationVote, a.pokerService), a.store, a.pokerService))
	a.mux.Handle(a.config.path.postViolationComplaint, middleware.NewAuthMiddleware(appHttp.NewPostViolationComplaintHandler(a.config.path.postViolationComplaint, a.pokerService), a.store, a.pokerService))

	a.mux.Handle(
		a.config.path.postViolationRequestVote,
		middleware.NewAuthMiddleware(
			appHttp.NewPostViolationRequestVoteHandler(
				a.config.path.postViolationRequestVote,
				a.pokerService,
			),
			a.store,
			a.pokerService,
		),
	)

	a.mux.Handle(
		a.config.path.postViolationRequestComplaint,
		middleware.NewAuthMiddleware(
			appHttp.NewPostViolationRequestComplaintHandler(
				a.config.path.postViolationRequestComplaint,
				a.pokerService,
			),
			a.store,
			a.pokerService,
		),
	)
	fmt.Println("start server")

	return a.server.ListenAndServe()
}

func NewApp(ctx context.Context, config config, dbConn *pgxpool.Pool) (*App, error) {

	var (
		mux   = http.NewServeMux()
		hub   = ws.NewHub()
		store = sessions.NewCookieStore([]byte(config.sectrets.storeSecret))
	)

	// Настраиваем опции для CookieStore для работы с мобильными приложениями
	// ВАЖНО: Secure должен быть false для HTTP, true для HTTPS
	// В продакшене с HTTPS установите Secure: true
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,            // 30 дней
		HttpOnly: true,                  // Разрешить доступ для мобильных приложений
		Secure:   true,                  // Для HTTP (в продакшене с HTTPS установить true)
		SameSite: http.SameSiteNoneMode, // Для кросс-доменных запросов (мобильное приложение)
	}

	// Build repository
	repo := repository.NewPokerRepository(dbConn)

	// Build token services
	accessTokenService := tokenservice.NewtokenService([]byte(config.sectrets.accessTokenSecret), 1*time.Hour, model.Access_Token_Type)
	refreshTokenService := tokenservice.NewtokenService([]byte(config.sectrets.refreshTokenSecret), 30*24*time.Hour, model.Refresh_Token_Type)

	// Build providers user data map from config
	providersMap := make(authinterface.ProvidersUserData)
	for key, prov := range config.provadersConf {
		if prov != nil && prov.ProviderUserData != nil {
			providersMap[key] = prov.ProviderUserData
		}
	}

	// Build service
	pokerService := service.NewPokerService(repo, &hubAdapter{Hub: hub}, accessTokenService, refreshTokenService, providersMap)

	// Build object storage uploader (Yandex Object Storage)
	uploader, err := objectstorage.NewUploader(
		config.sectrets.yosEndpoint,
		config.sectrets.yosAccessKey,
		config.sectrets.yosSecretKey,
		config.sectrets.yosBucket,
		config.sectrets.yosCdnBaseURL,
		true,
	)
	if err != nil {
		return nil, err
	}

	// Создаем CORS middleware для мобильного приложения
	corsOptions := cors.Options{
		// Добавляем все необходимые методы
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		// Разрешаем все стандартные заголовки + кастомные
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "X-CSRF-Token", "Custom-Header",
			"Cookie",             // Важно для работы с куками
			"X-Mobile-Signature", // HMAC-SHA256 подпись запроса (Base64)
			"X-Mobile-Timestamp", // Timestamp запроса (миллисекунды)
		},
		// Разрешаем куки и авторизацию
		AllowCredentials: true,
		// Разрешаем клиенту видеть заголовок Set-Cookie
		ExposedHeaders: []string{"Set-Cookie"},
		// Опционально: максимальное время кеширования preflight-запросов
		MaxAge: 86400,
	}

	// Создаем whitelist для веб-приложений
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range config.corsAllowedOrigins {
		allowedOriginsMap[origin] = true
	}

	// Дефолтные origins для разработки
	devOrigins := []string{
		"http://localhost:3000",
		"http://10.0.2.2",
	}
	devOriginsMap := make(map[string]bool)
	for _, origin := range devOrigins {
		devOriginsMap[origin] = true
	}

	// Используем AllowOriginVaryRequestFunc для доступа к заголовкам запроса
	corsOptions.AllowOriginVaryRequestFunc = func(r *http.Request, origin string) (bool, []string) {
		// Если origin пустой - разрешаем для мобильных приложений
		// Проверка подписи будет выполнена в MobileSignatureMiddleware
		if origin == "" {
			// Проверяем наличие заголовков подписи (проверка самой подписи в middleware)
			signature := r.Header.Get("X-Mobile-Signature")
			timestamp := r.Header.Get("X-Mobile-Timestamp")
			if signature != "" && timestamp != "" {
				// Мобильное приложение с заголовками подписи
				return true, []string{"X-Mobile-Signature", "X-Mobile-Timestamp"}
			}
			// Пустой origin без заголовков подписи - блокируем (защита от атак)
			return false, nil
		}

		// Для веб-приложений: строгая проверка whitelist
		if len(config.corsAllowedOrigins) > 0 {
			// В продакшене: проверяем только whitelist из конфига
			if allowedOriginsMap[origin] {
				return true, nil
			}
			return false, nil
		}

		// В режиме разработки: проверяем дефолтные origins
		if devOriginsMap[origin] {
			return true, nil
		}

		return false, nil
	}

	corsMiddleware := cors.New(corsOptions)

	// Обертываем основной обработчик: сначала CORS, потом логирование
	handler := corsMiddleware.Handler(
		middleware.NewLogMux(mux),
	)

	return &App{
		mux:           mux,
		server:        &http.Server{Addr: config.addr, Handler: handler, ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second},
		pokerService:  pokerService,
		config:        config,
		hub:           hub,
		store:         store,
		provadersConf: config.provadersConf,
		uploader:      uploader,
	}, nil

}
