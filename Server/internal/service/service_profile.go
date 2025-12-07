package service

import (
	"context"

	"github.com/inzarubin80/Server/internal/model"
)

// GetUserProfile возвращает профиль пользователя с базовой информацией и списком подключенных провайдеров.
func (s *PokerService) GetUserProfile(ctx context.Context, userID model.UserID) (*model.UserProfile, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	authProviders, err := s.repository.GetUserAuthProvidersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	connected := make([]string, 0, len(authProviders))
	for _, p := range authProviders {
		if p != nil && p.Provider != "" {
			connected = append(connected, p.Provider)
		}
	}

	return &model.UserProfile{
		ID:                 user.ID,
		Name:               user.Name,
		AvatarURL:          user.AvatarURL,
		BoostyURL:          user.BoostyURL,
		ConnectedProviders: connected,
	}, nil
}

// UpdateUserProfile обновляет отображаемое имя и ссылку на Boosty.
// Пустые поля в запросе не изменяют соответствующие значения.
func (s *PokerService) UpdateUserProfile(ctx context.Context, userID model.UserID, name *string, boostyURL *string) (*model.UserProfile, error) {
	if name != nil {
		if err := s.repository.SetUserName(ctx, userID, *name); err != nil {
			return nil, err
		}
	}

	if boostyURL != nil {
		if err := s.repository.UpdateUserBoostyURL(ctx, userID, boostyURL); err != nil {
			return nil, err
		}
	}

	return s.GetUserProfile(ctx, userID)
}


