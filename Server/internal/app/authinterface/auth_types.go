package authinterface

import (
	"context"

	"github.com/inzarubin80/Server/internal/model"

	"golang.org/x/oauth2"
)

type (
	TokenService interface {
		GenerateToken(userID model.UserID, isAdmin bool) (string, error)
		ValidateToken(tokenString string) (*model.Claims, error)
	}

	ProviderUserData interface {
		GetUserData(ctx context.Context, authorizationCode string, codeVerifier string) (*model.UserProfileFromProvider, error)
	}

	ProvidersUserData map[string]ProviderUserData

	ProviderOauthConf struct {
		Oauth2Config     *oauth2.Config
		UrlUserData      string
		IconSVG          string
		DisplayName      string
		ProviderUserData ProviderUserData // Добавлено поле для получения user data
	}

	MapProviderOauthConf map[string]*ProviderOauthConf

	ProviderOauthConfFrontend struct {
		Provider    string
		IconSVG     string
		Name string
	}
)
