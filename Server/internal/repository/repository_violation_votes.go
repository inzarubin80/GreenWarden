package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inzarubin80/Server/internal/model"
	sqlc_repository "github.com/inzarubin80/Server/internal/repository_sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// SetViolationVote upserts a user's vote for a violation.
func (r *Repository) SetViolationVote(ctx context.Context, violationID model.ViolationID, userID model.UserID, value string) (*model.ViolationVotes, error) {
	if value != "like" && value != "dislike" {
		return nil, fmt.Errorf("invalid vote value: %s", value)
	}

	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	arg := &sqlc_repository.UpsertViolationVoteParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		UserID:      int64(userID),
		Value:       value,
	}

	_, err = reposqlsc.UpsertViolationVote(ctx, arg)
	if err != nil {
		return nil, err
	}

	return r.GetViolationVotes(ctx, violationID, userID)
}

// SetViolationRequestVote upserts a user's vote for a violation that is made in the context of a specific request.
// It stores both violation_id and request_id in the violation_votes table.
func (r *Repository) SetViolationRequestVote(ctx context.Context, violationID model.ViolationID, requestID string, userID model.UserID, value string) (*model.ViolationVotes, error) {
	if value != "like" && value != "dislike" {
		return nil, fmt.Errorf("invalid vote value: %s", value)
	}

	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID: %w", err)
	}

	arg := &sqlc_repository.UpsertViolationRequestVoteParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		RequestID:   pgtype.UUID{Bytes: reqUUID, Valid: true},
		UserID:      int64(userID),
		Value:       value,
	}

	_, err = reposqlsc.UpsertViolationRequestVote(ctx, arg)
	if err != nil {
		return nil, err
	}

	return r.GetViolationRequestVotes(ctx, violationID, requestID, userID)
}

// GetViolationRequestVotes returns aggregated likes/dislikes and optional user vote for a specific request.
func (r *Repository) GetViolationRequestVotes(ctx context.Context, violationID model.ViolationID, requestID string, userID model.UserID) (*model.ViolationVotes, error) {
	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID: %w", err)
	}

	// Aggregate likes/dislikes and user_vote per request_id.
	row := r.conn.QueryRow(ctx, `
		SELECT
		    COALESCE(SUM(CASE WHEN value = 'like' THEN 1 ELSE 0 END), 0)   AS likes,
		    COALESCE(SUM(CASE WHEN value = 'dislike' THEN 1 ELSE 0 END), 0) AS dislikes,
		    COALESCE(
		        MAX(CASE WHEN user_id = $2 THEN value END),
		        ''
		    ) AS user_vote
		FROM violation_votes
		WHERE request_id = $1
	`, reqUUID, int64(userID))

	var (
		likes    int64
		dislikes int64
		userVote string
	)

	if err := row.Scan(&likes, &dislikes, &userVote); err != nil {
		// Если голосов еще нет, возвращаем нули без ошибки.
		if errors.Is(err, sql.ErrNoRows) {
			return &model.ViolationVotes{
				ViolationID: violationID,
				Likes:       0,
				Dislikes:    0,
				UserVote:    "",
			}, nil
		}
		return nil, err
	}

	return &model.ViolationVotes{
		ViolationID: violationID,
		Likes:       likes,
		Dislikes:    dislikes,
		UserVote:    userVote,
	}, nil
}

// DeleteViolationVote removes a user's vote for a violation.
func (r *Repository) DeleteViolationVote(ctx context.Context, violationID model.ViolationID, userID model.UserID) (*model.ViolationVotes, error) {
	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	arg := &sqlc_repository.DeleteViolationVoteParams{
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		UserID:      int64(userID),
	}

	if err := reposqlsc.DeleteViolationVote(ctx, arg); err != nil {
		return nil, err
	}

	return r.GetViolationVotes(ctx, violationID, userID)
}

// GetViolationVotes returns aggregated likes/dislikes and optional user vote.
func (r *Repository) GetViolationVotes(ctx context.Context, violationID model.ViolationID, userID model.UserID) (*model.ViolationVotes, error) {
	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	arg := &sqlc_repository.GetViolationVotesAggregatedParams{
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		UserID:      int64(userID),
	}

	row, err := reposqlsc.GetViolationVotesAggregated(ctx, arg)
	if err != nil {
		// Если голосов еще нет или violation не найден, возвращаем нули без ошибки.
		return &model.ViolationVotes{
			ViolationID: violationID,
			Likes:       0,
			Dislikes:    0,
			UserVote:    "",
		}, nil
	}

	likes, _ := row.Likes.(int64)
	dislikes, _ := row.Dislikes.(int64)
	userVoteStr, _ := row.UserVote.(string)

	return &model.ViolationVotes{
		ViolationID: violationID,
		Likes:       likes,
		Dislikes:    dislikes,
		UserVote:    userVoteStr,
	}, nil
}

// CreateViolationComplaint creates a complaint for a violation from a user.
func (r *Repository) CreateViolationComplaint(ctx context.Context, violationID model.ViolationID, userID model.UserID, reason, message string) (*model.ViolationComplaint, error) {
	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	complaintID := uuid.New()

	var (
		reasonPtr  *string
		messagePtr *string
	)
	if reason != "" {
		reasonPtr = &reason
	}
	if message != "" {
		messagePtr = &message
	}

	arg := &sqlc_repository.CreateViolationComplaintParams{
		ID:          pgtype.UUID{Bytes: complaintID, Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		UserID:      int64(userID),
		Reason:      reasonPtr,
		Message:     messagePtr,
	}

	dbCompl, err := reposqlsc.CreateViolationComplaint(ctx, arg)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if dbCompl.CreatedAt.Valid {
		createdAt = dbCompl.CreatedAt.Time
	}

	out := &model.ViolationComplaint{
		ID:          uuid.UUID(dbCompl.ID.Bytes).String(),
		ViolationID: violationID,
		UserID:      userID,
		CreatedAt:   createdAt,
	}
	if dbCompl.Reason != nil {
		out.Reason = *dbCompl.Reason
	}
	if dbCompl.Message != nil {
		out.Message = *dbCompl.Message
	}
	if dbCompl.RequestID.Valid {
		out.RequestID = uuid.UUID(dbCompl.RequestID.Bytes).String()
	}

	return out, nil
}

// CreateViolationRequestComplaint creates a complaint for a violation request from a user.
// It stores both violation_id and request_id in the violation_complaints table.
func (r *Repository) CreateViolationRequestComplaint(ctx context.Context, violationID model.ViolationID, requestID string, userID model.UserID, reason, message string) (*model.ViolationComplaint, error) {
	reposqlsc := sqlc_repository.New(r.conn)

	vUUID, err := uuid.Parse(string(violationID))
	if err != nil {
		return nil, fmt.Errorf("invalid violation ID: %w", err)
	}

	reqUUID, err := uuid.Parse(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid request ID: %w", err)
	}

	complaintID := uuid.New()

	var (
		reasonPtr  *string
		messagePtr *string
	)
	if reason != "" {
		reasonPtr = &reason
	}
	if message != "" {
		messagePtr = &message
	}

	arg := &sqlc_repository.CreateViolationRequestComplaintParams{
		ID:          pgtype.UUID{Bytes: complaintID, Valid: true},
		ViolationID: pgtype.UUID{Bytes: vUUID, Valid: true},
		RequestID:   pgtype.UUID{Bytes: reqUUID, Valid: true},
		UserID:      int64(userID),
		Reason:      reasonPtr,
		Message:     messagePtr,
	}

	dbCompl, err := reposqlsc.CreateViolationRequestComplaint(ctx, arg)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if dbCompl.CreatedAt.Valid {
		createdAt = dbCompl.CreatedAt.Time
	}

	out := &model.ViolationComplaint{
		ID:          uuid.UUID(dbCompl.ID.Bytes).String(),
		ViolationID: violationID,
		UserID:      userID,
		CreatedAt:   createdAt,
	}
	if dbCompl.Reason != nil {
		out.Reason = *dbCompl.Reason
	}
	if dbCompl.Message != nil {
		out.Message = *dbCompl.Message
	}
	if dbCompl.RequestID.Valid {
		out.RequestID = uuid.UUID(dbCompl.RequestID.Bytes).String()
	}

	return out, nil
}


