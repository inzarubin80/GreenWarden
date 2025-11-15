package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inzarubin80/Server/internal/model"
)

const (
	sqlInsertViolation = `
INSERT INTO violations (id, user_id, type, description, lat, lng, status, confirmations_count, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8, NOW(), NOW())
RETURNING id, user_id, type, description, lat, lng, status, confirmations_count, created_at, updated_at;
`
)

func (r *Repository) CreateViolation(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64) (*model.Violation, error) {
	id := uuid.New().String()
	row := r.conn.QueryRow(ctx, sqlInsertViolation, id, int64(userID), string(vType), description, lat, lng, "new", 0)

	var res model.Violation
	var createdAt time.Time
	var updatedAt time.Time

	if err := row.Scan(
		&res.ID,
		&res.UserID,
		&res.Type,
		&res.Description,
		&res.Lat,
		&res.Lng,
		&res.Status,
		&res.ConfirmationsCount,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	res.CreatedAt = createdAt
	res.UpdatedAt = updatedAt
	return &res, nil
}


