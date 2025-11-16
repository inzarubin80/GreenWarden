package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inzarubin80/Server/internal/model"
	sqlc_repository "github.com/inzarubin80/Server/internal/repository_sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Repository) CreateViolation(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64) (*model.Violation, error) {
	reposqlsc := sqlc_repository.New(r.conn)
	violationUUID := uuid.New()
	var descPtr *string
	if description != "" {
		descPtr = &description
	}
	arg := &sqlc_repository.CreateViolationParams{
		ID:          pgtype.UUID{Bytes: violationUUID, Valid: true},
		UserID:      int64(userID),
		Type:        string(vType),
		Description: descPtr,
		Lat:         lat,
		Lng:         lng,
	}
	v, err := reposqlsc.CreateViolation(ctx, arg)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	updatedAt := time.Now()
	// Prefer DB timestamps when present
	if v.CreatedAt.Valid {
		createdAt = v.CreatedAt.Time
	}
	if v.UpdatedAt.Valid {
		updatedAt = v.UpdatedAt.Time
	}
	desc := ""
	if v.Description != nil {
		desc = *v.Description
	}

	return &model.Violation{
		ID:                 model.ViolationID(violationUUID.String()),
		UserID:             model.UserID(v.UserID),
		Type:               model.ViolationType(v.Type),
		Description:        desc,
		Lat:                v.Lat,
		Lng:                v.Lng,
		Status:             model.ViolationStatus(v.Status),
		ConfirmationsCount: int(v.ConfirmationsCount),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func (r *Repository) AddViolationPhoto(ctx context.Context, violationID string, url string, thumbURL string) (*model.ViolationPhoto, error) {
	reposqlsc := sqlc_repository.New(r.conn)
	photoUUID := uuid.New()
	vUUID, err := uuid.Parse(violationID)
	if err != nil {
		return nil, err
	}
	var thumbPtr *string
	if thumbURL != "" {
		thumbPtr = &thumbURL
	}
	arg := &sqlc_repository.AddViolationPhotoParams{
		ID:          pgtype.UUID{Bytes: photoUUID, Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		Url:         url,
		ThumbUrl:    thumbPtr,
	}
	p, err := reposqlsc.AddViolationPhoto(ctx, arg)
	if err != nil {
		return nil, err
	}
	thumb := ""
	if p.ThumbUrl != nil {
		thumb = *p.ThumbUrl
	}
	return &model.ViolationPhoto{
		ID:           photoUUID.String(),
		ViolationID:  vUUID.String(),
		URL:          p.Url,
		ThumbnailURL: thumb,
	}, nil
}
