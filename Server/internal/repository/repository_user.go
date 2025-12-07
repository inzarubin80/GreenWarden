package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inzarubin80/Server/internal/model"
	sqlc_repository "github.com/inzarubin80/Server/internal/repository_sqlc"
)

func (r *Repository) CreateUser(ctx context.Context, userData *model.UserProfileFromProvider) (*model.User, error) {

	reposqlsc := sqlc_repository.New(r.conn)
	userID, err := reposqlsc.CreateUser(ctx, userData.Name)

	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:   model.UserID(userID),
		Name: userData.Name,
	}

	return user, nil
}

func (r *Repository) SetUserName(ctx context.Context, userID model.UserID, name string) error {

	reposqlsc := sqlc_repository.New(r.conn)
	arg := &sqlc_repository.UpdateUserNameParams{
		Name:   name,
		UserID: int64(userID),
	}
	_, err := reposqlsc.UpdateUserName(ctx, arg)

	return err

}

func (r *Repository) SetUserAvatarIfEmpty(ctx context.Context, userID model.UserID, avatarURL *string) error {
	reposqlsc := sqlc_repository.New(r.conn)
	arg := &sqlc_repository.SetUserAvatarIfEmptyParams{
		UserID:    int64(userID),
		AvatarUrl: avatarURL,
	}
	return reposqlsc.SetUserAvatarIfEmpty(ctx, arg)
}

func (r *Repository) UpdateUserBoostyURL(ctx context.Context, userID model.UserID, boostyURL *string) error {
	reposqlsc := sqlc_repository.New(r.conn)
	arg := &sqlc_repository.UpdateUserBoostyURLParams{
		UserID:    int64(userID),
		BoostyUrl: boostyURL,
	}
	return reposqlsc.UpdateUserBoostyURL(ctx, arg)
}

func (r *Repository) UpdateUserAvatar(ctx context.Context, userID model.UserID, avatarURL *string) error {
	reposqlsc := sqlc_repository.New(r.conn)
	arg := &sqlc_repository.UpdateUserAvatarParams{
		UserID:    int64(userID),
		AvatarUrl: avatarURL,
	}
	return reposqlsc.UpdateUserAvatar(ctx, arg)
}



func (r *Repository) GetUser(ctx context.Context, userID model.UserID) (*model.User, error) {

	reposqlsc := sqlc_repository.New(r.conn)
	user, err := reposqlsc.GetUserByID(ctx, int64(userID))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", model.ErrorNotFound, err)
		}
		return nil, err
	}

	res := &model.User{
		ID:   model.UserID(user.UserID),
		Name: user.Name,
	}
	if user.EvaluationStrategy != nil {
		res.EvaluationStrategy = *user.EvaluationStrategy
	}
	if user.MaximumScore != nil {
		res.MaximumScore = int(*user.MaximumScore)
	}
	if user.AvatarUrl != nil {
		res.AvatarURL = *user.AvatarUrl
	}
	if user.BoostyUrl != nil {
		res.BoostyURL = *user.BoostyUrl
	}

	return res, nil

}

func (r *Repository) GetUsersByIDs(ctx context.Context, userIDs []model.UserID) ([]*model.User, error) {

	reposqlsc := sqlc_repository.New(r.conn)
	arg := make([]int64, len(userIDs), len(userIDs))
	for i, value := range userIDs {
		arg[i] = int64(value)
	}

	users, err := reposqlsc.GetUsersByIDs(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", model.ErrorNotFound, err)
		}
		return nil, err
	}

	usersRes := make([]*model.User, len(users))

	for i, value := range users {
		u := &model.User{
			ID:   model.UserID(value.UserID),
			Name: value.Name,
		}
		if value.EvaluationStrategy != nil {
			u.EvaluationStrategy = *value.EvaluationStrategy
		}
		if value.MaximumScore != nil {
			u.MaximumScore = int(*value.MaximumScore)
		}
		if value.AvatarUrl != nil {
			u.AvatarURL = *value.AvatarUrl
		}
		if value.BoostyUrl != nil {
			u.BoostyURL = *value.BoostyUrl
		}
		usersRes[i] = u
	}

	return usersRes, nil

}

