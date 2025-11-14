package service

import (
	"context"

	authinterface "github.com/inzarubin80/Server/internal/app/authinterface"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	PokerService struct {
		repository          Repository
		hub                 Hub
		accessTokenService  TokenService
		refreshTokenService TokenService
		providersUserData   authinterface.ProvidersUserData
	}

	Repository interface {

		//User
		GetUserAuthProvidersByProviderUid(ctx context.Context, ProviderUid string, Provider string) (*model.UserAuthProviders, error)
		AddUserAuthProviders(ctx context.Context, userProfileFromProvide *model.UserProfileFromProvider, userID model.UserID) (*model.UserAuthProviders, error)
		CreateUser(ctx context.Context, userData *model.UserProfileFromProvider) (*model.User, error)
		GetUsersByIDs(ctx context.Context, userIDs []model.UserID) ([]*model.User, error)
		GetUserIDsByPokerID(ctx context.Context, pokerID model.PokerID) ([]model.UserID, error)
		SetUserName(ctx context.Context, userID model.UserID, name string) error
		GetUser(ctx context.Context, userID model.UserID) (*model.User, error)
	}

	TokenService interface {
		GenerateToken(userID model.UserID) (string, error)
		ValidateToken(tokenString string) (*model.Claims, error)
	}

	ProviderUserData interface {
		GetUserData(ctx context.Context, authorizationCode string) (*model.UserProfileFromProvider, error)
	}

	Hub interface {
		AddMessage(pokerID model.PokerID, payload any) error
		AddMessageForUser(pokerID model.PokerID, userID model.UserID, payload any) error
		GetActiveUsersID(pokerID model.PokerID) ([]model.UserID, error)
	}
)

func NewPokerService(repository Repository, hub Hub, accessTokenService TokenService, refreshTokenService TokenService, providersUserData authinterface.ProvidersUserData) *PokerService {
	return &PokerService{
		repository:          repository,
		hub:                 hub,
		accessTokenService:  accessTokenService,
		refreshTokenService: refreshTokenService,
		providersUserData:   providersUserData,
	}
}
