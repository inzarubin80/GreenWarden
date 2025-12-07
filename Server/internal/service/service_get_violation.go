package service

import (
	"context"
	"errors"

	"github.com/inzarubin80/Server/internal/model"
)

func (s *PokerService) GetViolationByID(ctx context.Context, id model.ViolationID, userID model.UserID) (*model.Violation, error) {
	// Validate UUID format
	if len(string(id)) == 0 {
		return nil, ErrBadRequest("violation ID is required")
	}
	
	violation, err := s.repository.GetViolationByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, model.ErrorNotFound) {
			return nil, model.ErrorNotFound
		}
		return nil, err
	}
	
	return violation, nil
}

