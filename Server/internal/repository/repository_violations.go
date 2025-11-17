package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

func (r *Repository) ListViolations(ctx context.Context, f *model.ListViolationsFilters) ([]*model.Violation, int64, error) {
	// Build dynamic SQL to avoid dependency on regenerated sqlc code
	where := "WHERE 1=1"
	args := []interface{}{}

	if f.Type != nil && *f.Type != "" {
		where += " AND type = $" + strconv.Itoa(len(args)+1)
		args = append(args, *f.Type)
	}
	if f.Status != nil && *f.Status != "" {
		where += " AND status = $" + strconv.Itoa(len(args)+1)
		args = append(args, *f.Status)
	}
	if f.From != nil {
		where += " AND created_at >= $" + strconv.Itoa(len(args)+1)
		args = append(args, *f.From)
	}
	if f.To != nil {
		where += " AND created_at <= $" + strconv.Itoa(len(args)+1)
		args = append(args, *f.To)
	}
	if f.MinLng != nil && f.MinLat != nil && f.MaxLng != nil && f.MaxLat != nil {
		where += " AND lng BETWEEN $" + strconv.Itoa(len(args)+1) + " AND $" + strconv.Itoa(len(args)+2)
		args = append(args, *f.MinLng, *f.MaxLng)
		where += " AND lat BETWEEN $" + strconv.Itoa(len(args)+1) + " AND $" + strconv.Itoa(len(args)+2)
		args = append(args, *f.MinLat, *f.MaxLat)
	}

	limit := f.PageSize
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	countSQL := "SELECT count(1) FROM violations " + where
	row := r.conn.QueryRow(ctx, countSQL, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL := "SELECT id::text, user_id, type, COALESCE(description,''), lat, lng, status, confirmations_count, created_at, updated_at FROM violations " + where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, int32(limit), int32(offset))
	rows, err := r.conn.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*model.Violation
	for rows.Next() {
		var (
			idStr        string
			userID       int64
			vType        string
			desc         string
			lat, lng     float64
			status       string
			confirmCnt   int32
			createdAt    time.Time
			updatedAt    time.Time
		)
		if err := rows.Scan(&idStr, &userID, &vType, &desc, &lat, &lng, &status, &confirmCnt, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &model.Violation{
			ID:                 model.ViolationID(idStr),
			UserID:             model.UserID(userID),
			Type:               model.ViolationType(vType),
			Description:        desc,
			Lat:                lat,
			Lng:                lng,
			Status:             model.ViolationStatus(status),
			ConfirmationsCount: int(confirmCnt),
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		})
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return out, total, nil
}

func (r *Repository) GetViolationByID(ctx context.Context, id model.ViolationID) (*model.Violation, error) {
	reposqlsc := sqlc_repository.New(r.conn)
	
	violationUUID, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}
	
	v, err := reposqlsc.GetViolationByID(ctx, pgtype.UUID{Bytes: violationUUID, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: violation not found", model.ErrorNotFound)
		}
		return nil, err
	}
	
	// Get photos for this violation
	photos, err := reposqlsc.GetPhotosByViolationID(ctx, pgtype.UUID{Bytes: violationUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	
	// Map photos
	violationPhotos := make([]model.ViolationPhoto, 0, len(photos))
	for _, p := range photos {
		thumb := ""
		if p.ThumbUrl != nil {
			thumb = *p.ThumbUrl
		}
		violationPhotos = append(violationPhotos, model.ViolationPhoto{
			ID:           uuid.UUID(p.ID.Bytes).String(),
			ViolationID:  uuid.UUID(p.ViolationID.Bytes).String(),
			URL:          p.Url,
			ThumbnailURL: thumb,
		})
	}
	
	createdAt := time.Now()
	updatedAt := time.Now()
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
		Photos:             violationPhotos,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

