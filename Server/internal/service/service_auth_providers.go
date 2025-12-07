package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/inzarubin80/Server/internal/model"
)

var (
	ErrProviderAlreadyLinkedToAnotherUser = errors.New("provider account is already linked to another user")
	ErrLastAuthProvider                   = errors.New("cannot unlink the last authentication provider")
	ErrProviderNotLinked                  = errors.New("provider is not linked to this user")
)

// LinkAuthProvider привязывает новый OAuth-провайдер к уже авторизованному пользователю.
// Флоу: мобилка получает authorization code + code_verifier от провайдера и шлёт их сюда.
func (s *PokerService) LinkAuthProvider(ctx context.Context, userID model.UserID, providerKey, authorizationCode, codeVerifier string) (*model.UserProfile, error) {
	provider, ok := s.providersUserData[providerKey]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}

	userProfileFromProvider, err := provider.GetUserData(ctx, authorizationCode, codeVerifier)
	if err != nil {
		return nil, err
	}

	// Проверяем, не привязан ли уже этот внешний аккаунт к какому-то пользователю
	existing, err := s.repository.GetUserAuthProvidersByProviderUid(ctx, userProfileFromProvider.ProviderID, userProfileFromProvider.ProviderName)
	if err != nil && !errors.Is(err, model.ErrorNotFound) {
		return nil, err
	}

	if existing != nil {
		if existing.UserID != userID {
			return nil, ErrProviderAlreadyLinkedToAnotherUser
		}
		// Уже привязан к этому пользователю — считаем операцию идемпотентной
		return s.GetUserProfile(ctx, userID)
	}

	// Привязываем провайдера к текущему пользователю
	if _, err := s.repository.AddUserAuthProviders(ctx, userProfileFromProvider, userID); err != nil {
		return nil, err
	}

	// Если у пользователя ещё нет аватара, попробуем установить его из провайдера
	if userProfileFromProvider.AvatarURL != "" {
		if err := s.repository.SetUserAvatarIfEmpty(ctx, userID, &userProfileFromProvider.AvatarURL); err != nil {
			// Не проваливаем операцию привязки из-за аватара
			_ = err
		}
	}

	return s.GetUserProfile(ctx, userID)
}

// UnlinkAuthProvider отвязывает провайдера от пользователя.
// Нельзя отвязать последний способ входа.
func (s *PokerService) UnlinkAuthProvider(ctx context.Context, userID model.UserID, providerKey string) (*model.UserProfile, error) {
	// Получаем все привязки пользователя
	providers, err := s.repository.GetUserAuthProvidersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(providers) == 0 {
		return nil, ErrProviderNotLinked
	}

	// Проверяем, что такой провайдер вообще есть у пользователя
	found := false
	for _, p := range providers {
		if p.Provider == providerKey {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrProviderNotLinked
	}

	// Если это единственный провайдер — запрещаем отвязку
	if len(providers) == 1 {
		return nil, ErrLastAuthProvider
	}

	if err := s.repository.DeleteUserAuthProvider(ctx, userID, providerKey); err != nil {
		return nil, err
	}

	return s.GetUserProfile(ctx, userID)
}


