package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inzarubin80/Server/internal/model"
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

	query := `
		INSERT INTO violation_chat_messages (id, violation_id, user_id, text, is_system)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	var (
		createdAt time.Time
		updatedAt *time.Time
	)
	if err := r.conn.QueryRow(ctx, query, msgUUID, vUUID, int64(userID), text, isSystem).Scan(&createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return &model.ViolationChatMessage{
		ID:          msgUUID.String(),
		ViolationID: violationID,
		UserID:      userID,
		Text:        text,
		IsSystem:    isSystem,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
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

	// count total
	var total int64
	countSQL := `SELECT count(1) FROM violation_chat_messages WHERE violation_id = $1`
	if err := r.conn.QueryRow(ctx, countSQL, vUUID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	listSQL := `
		SELECT id, violation_id, user_id, text, is_system, created_at, updated_at
		FROM violation_chat_messages
		WHERE violation_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.conn.Query(ctx, listSQL, vUUID, int32(pageSize), int32(offset))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*model.ViolationChatMessage
	for rows.Next() {
		var (
			idUUID    uuid.UUID
			vIDUUID   uuid.UUID
			uID       int64
			text      string
			isSystem  bool
			created   time.Time
			updatedAt *time.Time
		)

		if err := rows.Scan(&idUUID, &vIDUUID, &uID, &text, &isSystem, &created, &updatedAt); err != nil {
			return nil, 0, err
		}

		out = append(out, &model.ViolationChatMessage{
			ID:          idUUID.String(),
			ViolationID: violationID,
			UserID:      model.UserID(uID),
			Text:        text,
			IsSystem:    isSystem,
			CreatedAt:   created,
			UpdatedAt:   updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
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

	query := `
		UPDATE violation_chat_messages
		SET text = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND is_system = FALSE
		RETURNING id, violation_id, user_id, text, is_system, created_at, updated_at
	`

	var (
		idUUID    uuid.UUID
		vIDUUID   uuid.UUID
		uID       int64
		isSystem  bool
		createdAt time.Time
		updatedAt *time.Time
	)

	if err := r.conn.QueryRow(ctx, query, text, msgUUID, int64(userID)).Scan(
		&idUUID, &vIDUUID, &uID, &text, &isSystem, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	return &model.ViolationChatMessage{
		ID:          idUUID.String(),
		ViolationID: model.ViolationID(vIDUUID.String()),
		UserID:      model.UserID(uID),
		Text:        text,
		IsSystem:    isSystem,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// DeleteViolationChatMessage deletes a message created by the given user and returns its violationID.
func (r *Repository) DeleteViolationChatMessage(ctx context.Context, messageID string, userID model.UserID) (model.ViolationID, error) {
	msgUUID, err := uuid.Parse(messageID)
	if err != nil {
		return "", fmt.Errorf("invalid message ID: %w", err)
	}

	query := `
		DELETE FROM violation_chat_messages
		WHERE id = $1 AND user_id = $2 AND is_system = FALSE
		RETURNING violation_id
	`

	var vID uuid.UUID
	if err := r.conn.QueryRow(ctx, query, msgUUID, int64(userID)).Scan(&vID); err != nil {
		return "", err
	}

	return model.ViolationID(vID.String()), nil
}


