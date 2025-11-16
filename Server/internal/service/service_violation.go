package service

import (
	"context"
	"fmt"
	"mime/multipart"

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

// CreateViolationWithPhotos creates violation and attaches uploaded photos (uploader and limits are expected to be wired in service; implementation can be extended later).
func (s *PokerService) CreateViolationWithPhotos(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.Violation, error) {
	violation, err := s.CreateViolation(ctx, userID, vType, description, lat, lng)
	if err != nil {
		return nil, err
	}
	added := 0
	for _, fh := range files {
		if maxPhotos > 0 && added >= maxPhotos {
			break
		}
		f, err := fh.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()

		key := fmt.Sprintf("violations/%s/%s", violation.ID, fh.Filename)
		url, err := upload(ctx, key, f, fh.Size, fh.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		p, err := s.repository.AddViolationPhoto(ctx, string(violation.ID), url, "")
		if err != nil {
			return nil, err
		}
		violation.Photos = append(violation.Photos, *p)
		added++
	}
	return violation, nil
}
