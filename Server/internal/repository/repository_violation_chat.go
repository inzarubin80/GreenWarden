package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inzarubin80/Server/internal/model"
	sqlc_repository "github.com/inzarubin80/Server/internal/repository_sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateViolationChatMessage creates a new chat message for a violation.
func (r *Repository) CreateViolationChatMessage(ctx context.Context, violationID model.ViolationID, userID model.UserID, text string, isSystem bool) (*model.ViolationChatMessage, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	msgUUID := uuid.New()

	reposqlsc := sqlc_repository.New(r.conn)

	arg := &sqlc_repository.CreateViolationChatMessageParams{
		ID:          pgtype.UUID{Bytes: msgUUID, Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		UserID:      int64(userID),
		Text:        text,
		IsSystem:    isSystem,
	}

	dbMsg, err := reposqlsc.CreateViolationChatMessage(ctx, arg)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if dbMsg.CreatedAt.Valid {
		createdAt = dbMsg.CreatedAt.Time
	}
	var updatedAtPtr *time.Time
	if dbMsg.UpdatedAt.Valid {
		t := dbMsg.UpdatedAt.Time
		updatedAtPtr = &t
	}

	return &model.ViolationChatMessage{
		ID:          uuid.UUID(dbMsg.ID.Bytes).String(),
		ViolationID: model.ViolationID(uuid.UUID(dbMsg.ViolationID.Bytes).String()),
		UserID:      model.UserID(dbMsg.UserID),
		Text:        dbMsg.Text,
		IsSystem:    dbMsg.IsSystem,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAtPtr,
	}, nil
}

// ListViolationChatMessages returns paginated messages for a violation ordered by created_at ASC.
func (r *Repository) ListViolationChatMessages(ctx context.Context, violationID model.ViolationID, page, pageSize int) ([]*model.ViolationChatMessage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid violation ID: %w", err)
	}

	// count total (простая агрегация, без sqlc)
	var total int64
	countSQL := `SELECT count(1) FROM violation_chat_messages WHERE violation_id = $1`
	if err := r.conn.QueryRow(ctx, countSQL, vUUID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	reposqlsc := sqlc_repository.New(r.conn)

	dbMsgs, err := reposqlsc.ListViolationChatMessages(ctx, &sqlc_repository.ListViolationChatMessagesParams{
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		Limit:       int32(pageSize),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	var out []*model.ViolationChatMessage
	for _, m := range dbMsgs {
		if m == nil {
			continue
		}
		createdAt := time.Now()
		if m.CreatedAt.Valid {
			createdAt = m.CreatedAt.Time
		}
		var updatedAtPtr *time.Time
		if m.UpdatedAt.Valid {
			t := m.UpdatedAt.Time
			updatedAtPtr = &t
		}
		out = append(out, &model.ViolationChatMessage{
			ID:          uuid.UUID(m.ID.Bytes).String(),
			ViolationID: model.ViolationID(uuid.UUID(m.ViolationID.Bytes).String()),
			UserID:      model.UserID(m.UserID),
			Text:        m.Text,
			IsSystem:    m.IsSystem,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAtPtr,
		})
	}

	return out, total, nil
}

// UpdateViolationChatMessage updates text of a message created by the given user.
func (r *Repository) UpdateViolationChatMessage(ctx context.Context, messageID string, userID model.UserID, text string) (*model.ViolationChatMessage, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}
	msgUUID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	reposqlsc := sqlc_repository.New(r.conn)

	arg := &sqlc_repository.UpdateViolationChatMessageParams{
		Text:   text,
		ID:     pgtype.UUID{Bytes: msgUUID, Valid: true},
		UserID: int64(userID),
	}

	dbMsg, err := reposqlsc.UpdateViolationChatMessage(ctx, arg)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if dbMsg.CreatedAt.Valid {
		createdAt = dbMsg.CreatedAt.Time
	}
	var updatedAtPtr *time.Time
	if dbMsg.UpdatedAt.Valid {
		t := dbMsg.UpdatedAt.Time
		updatedAtPtr = &t
	}

	return &model.ViolationChatMessage{
		ID:          uuid.UUID(dbMsg.ID.Bytes).String(),
		ViolationID: model.ViolationID(uuid.UUID(dbMsg.ViolationID.Bytes).String()),
		UserID:      model.UserID(dbMsg.UserID),
		Text:        dbMsg.Text,
		IsSystem:    dbMsg.IsSystem,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAtPtr,
	}, nil
}

// DeleteViolationChatMessage deletes a message created by the given user and returns its violationID.
func (r *Repository) DeleteViolationChatMessage(ctx context.Context, messageID string, userID model.UserID) (model.ViolationID, error) {
	msgUUID, err := uuid.Parse(messageID)
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %w", err)
	}

	reposqlsc := sqlc_repository.New(r.conn)

	arg := &sqlc_repository.DeleteViolationChatMessageParams{
		ID:     pgtype.UUID{Bytes: msgUUID, Valid: true},
		UserID: int64(userID),
	}

	vID, err := reposqlsc.DeleteViolationChatMessage(ctx, arg)
	if err != nil {
		return "", err
	}

	return model.ViolationID(uuid.UUID(vID.Bytes).String()), nil
}
