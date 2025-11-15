package service

import (
	"context"
	"fmt"

	"github.com/inzarubin80/Server/internal/model"
)

func (s *PokerService) CreateViolation(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64) (*model.Violation, error) {
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("invalid lat")
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid lng")
	}
	switch vType {
	case "garbage", "pollution", "air", "deforestation", "other":
	default:
		return nil, fmt.Errorf("invalid type")
	}

	return s.repository.CreateViolation(ctx, userID, vType, description, lat, lng)
}


