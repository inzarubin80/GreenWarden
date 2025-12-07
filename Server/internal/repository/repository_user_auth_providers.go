package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inzarubin80/Server/internal/model"
	sqlc_repository "github.com/inzarubin80/Server/internal/repository_sqlc"
)

func (r *Repository) GetUserAuthProvidersByProviderUid(ctx context.Context, ProviderUid string, Provider string) (*model.UserAuthProviders, error) {

	reposqlsc := sqlc_repository.New(r.conn)

	arg := &sqlc_repository.GetUserAuthProvidersByProviderUidParams{
		ProviderUid: ProviderUid,
		Provider:    Provider,
	}

	UserAuthProvider, err := reposqlsc.GetUserAuthProvidersByProviderUid(ctx, arg)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", model.ErrorNotFound, err)
		}
		return nil, err
	}

	res := &model.UserAuthProviders{
		UserID:      model.UserID(UserAuthProvider.UserID),
		ProviderUid: UserAuthProvider.ProviderUid,
		Provider:    UserAuthProvider.Provider,
	}
	if UserAuthProvider.Name != nil {
		res.Name = *UserAuthProvider.Name
	}
	return res, nil

}

func (r *Repository) AddUserAuthProviders(ctx context.Context, userProfileFromProvide *model.UserProfileFromProvider, userID model.UserID) (*model.UserAuthProviders, error) {

	reposqlsc := sqlc_repository.New(r.conn)

	arg := &sqlc_repository.AddUserAuthProvidersParams{
		UserID:      int64(userID),
		ProviderUid: userProfileFromProvide.ProviderID,
		Provider:    userProfileFromProvide.ProviderName,
		Name:        &userProfileFromProvide.Name,
	}

	UserAuthProvider, err := reposqlsc.AddUserAuthProviders(ctx, arg)

	if err != nil {
		return nil, err
	}

	res := &model.UserAuthProviders{
		UserID:      model.UserID(UserAuthProvider.UserID),
		ProviderUid: UserAuthProvider.ProviderUid,
		Provider:    UserAuthProvider.Provider,
	}
	if UserAuthProvider.Name != nil {
		res.Name = *UserAuthProvider.Name
	}
	return res, nil

}

func (r *Repository) DeleteUserAuthProvider(ctx context.Context, userID model.UserID, provider string) error {
	reposqlsc := sqlc_repository.New(r.conn)
	arg := &sqlc_repository.DeleteUserAuthProviderParams{
		UserID:   int64(userID),
		Provider: provider,
	}
	return reposqlsc.DeleteUserAuthProvider(ctx, arg)
}

func (r *Repository) GetUserAuthProvidersByUserID(ctx context.Context, userID model.UserID) ([]*model.UserAuthProviders, error) {
	reposqlsc := sqlc_repository.New(r.conn)

	rows, err := reposqlsc.GetUserAuthProvidersByUserID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*model.UserAuthProviders{}, nil
		}
		return nil, err
	}

	result := make([]*model.UserAuthProviders, 0, len(rows))
	for _, row := range rows {
		item := &model.UserAuthProviders{
			UserID:      model.UserID(row.UserID),
			ProviderUid: row.ProviderUid,
			Provider:    row.Provider,
		}
		if row.Name != nil {
			item.Name = *row.Name
		}
		result = append(result, item)
	}

	return result, nil

}
