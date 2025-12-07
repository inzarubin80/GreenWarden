package service

import (
	"context"
	"fmt"
	"log"

	"github.com/inzarubin80/Server/internal/model"
)

// SetViolationVote sets or changes a user's vote for a violation.
// value: "like" | "dislike" | "none".
func (s *PokerService) SetViolationVote(ctx context.Context, userID model.UserID, violationID model.ViolationID, value string) (*model.ViolationVotes, error) {
	if value != "like" && value != "dislike" && value != "none" {
		return nil, fmt.Errorf("invalid vote value: %s", value)
	}

	// Ensure violation exists.
	if _, err := s.repository.GetViolationByID(ctx, violationID, 0); err != nil {
		return nil, err
	}

	var (
		votes *model.ViolationVotes
		err   error
	)

	if value == "none" {
		votes, err = s.repository.DeleteViolationVote(ctx, violationID, userID)
	} else {
		votes, err = s.repository.SetViolationVote(ctx, violationID, userID, value)
	}
	if err != nil {
		return nil, err
	}

	// Broadcast WS event about updated counts.
	if votes != nil {
		event := map[string]any{
			"type": "violation_vote_updated",
			"payload": map[string]any{
				"violation_id": votes.ViolationID,
				"likes":        votes.Likes,
				"dislikes":     votes.Dislikes,
			},
		}
		_ = s.hub.BroadcastViolationMessage(votes.ViolationID, event)
	}

	return votes, nil
}

// SetViolationRequestVote sets or changes a user's vote for a violation request.
// It resolves the underlying violation by request ID and stores both violation_id and request_id.
func (s *PokerService) SetViolationRequestVote(ctx context.Context, userID model.UserID, requestID string, value string) (*model.ViolationVotes, error) {
	if value != "like" && value != "dislike" && value != "none" {
		return nil, fmt.Errorf("invalid vote value: %s", value)
	}

	// Load request to get underlying violation ID.
	req, err := s.repository.GetViolationRequestByID(ctx, requestID)
	if err != nil {
		log.Printf("[SetViolationRequestVote] failed to load request_id=%s for user_id=%d: %v", requestID, userID, err)
		return nil, err
	}

	// For "none" we simply remove the vote by violation ID, as before.
	if value == "none" {
		votes, err := s.repository.DeleteViolationVote(ctx, req.ViolationID, userID)
		if err != nil {
			return nil, err
		}

		// Broadcast WS event about updated counts.
		if votes != nil {
			event := map[string]any{
				"type": "violation_vote_updated",
				"payload": map[string]any{
					"violation_id": votes.ViolationID,
					"likes":        votes.Likes,
					"dislikes":     votes.Dislikes,
				},
			}
			_ = s.hub.BroadcastViolationMessage(votes.ViolationID, event)
		}

		return votes, nil
	}

	// For like/dislike we store both violation_id and request_id.
	votes, err := s.repository.SetViolationRequestVote(ctx, req.ViolationID, requestID, userID, value)
	if err != nil {
		return nil, err
	}

	// Broadcast WS event about updated counts.
	if votes != nil {
		event := map[string]any{
			"type": "violation_vote_updated",
			"payload": map[string]any{
				"violation_id": votes.ViolationID,
				"likes":        votes.Likes,
				"dislikes":     votes.Dislikes,
			},
		}
		_ = s.hub.BroadcastViolationMessage(votes.ViolationID, event)
	}

	return votes, nil
}

// GetViolationVotes returns aggregated likes/dislikes and, if userID != 0, user's vote.
func (s *PokerService) GetViolationVotes(ctx context.Context, userID model.UserID, violationID model.ViolationID) (*model.ViolationVotes, error) {
	// Не требуем существования violation здесь; репозиторий вернёт нули, если голосов нет.
	return s.repository.GetViolationVotes(ctx, violationID, userID)
}

// CreateViolationComplaint creates a complaint for a violation from a user.
func (s *PokerService) CreateViolationComplaint(ctx context.Context, userID model.UserID, violationID model.ViolationID, reason, message string) (*model.ViolationComplaint, error) {
	// Ensure violation exists.
	if _, err := s.repository.GetViolationByID(ctx, violationID, 0); err != nil {
		return nil, err
	}

	// TODO: добавить rate limiting/проверку повторов жалоб при необходимости.
	return s.repository.CreateViolationComplaint(ctx, violationID, userID, reason, message)
}

// CreateViolationRequestComplaint creates a complaint for a violation request.
// It resolves the underlying violation by request ID and stores both violation_id and request_id.
func (s *PokerService) CreateViolationRequestComplaint(ctx context.Context, userID model.UserID, requestID string, reason, message string) (*model.ViolationComplaint, error) {
	// Load request to get underlying violation ID.
	req, err := s.repository.GetViolationRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Ensure violation exists (same semantics as CreateViolationComplaint).
	if _, err := s.repository.GetViolationByID(ctx, req.ViolationID, 0); err != nil {
		return nil, err
	}

	return s.repository.CreateViolationRequestComplaint(ctx, req.ViolationID, requestID, userID, reason, message)
}


