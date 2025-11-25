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

// CreateViolationWithPhotos creates violation and attaches uploaded photos to the open request
func (s *PokerService) CreateViolationWithPhotos(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.Violation, error) {
	violation, err := s.CreateViolation(ctx, userID, vType, description, lat, lng)
	if err != nil {
		return nil, err
	}
	
	// Get the open request (created automatically)
	openRequest, err := s.repository.GetOpenRequestByViolationID(ctx, violation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open request: %w", err)
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

		// Upload to violation_requests/{request_id}/{filename}
		key := fmt.Sprintf("violation_requests/%s/%s", openRequest.ID, fh.Filename)
		url, err := upload(ctx, key, f, fh.Size, fh.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		
		// Add photo to the request (not to violation directly)
		p, err := s.repository.AddRequestPhoto(ctx, openRequest.ID, url, "")
		if err != nil {
			return nil, err
		}
		openRequest.Photos = append(openRequest.Photos, *p)
		added++
	}
	
	// Reload violation with requests to include photos
	violation, err = s.repository.GetViolationByID(ctx, violation.ID)
	if err != nil {
		return nil, err
	}
	
	return violation, nil
}

// CreateViolationRequestWithPhotos creates a request for closing a violation with photos
func (s *PokerService) CreateViolationRequestWithPhotos(ctx context.Context, violationID model.ViolationID, status model.ViolationRequestStatus, userID model.UserID, comment string, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.ViolationRequest, error) {
	// Validate status
	if status != "partially_closed" && status != "closed" {
		return nil, fmt.Errorf("invalid status: must be 'partially_closed' or 'closed'")
	}

	// Create request
	request, err := s.repository.CreateViolationRequest(ctx, violationID, status, userID, comment)
	if err != nil {
		return nil, err
	}

	// Add photos if provided
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

		// Upload to violation_requests/{request_id}/{filename}
		key := fmt.Sprintf("violation_requests/%s/%s", request.ID, fh.Filename)
		url, err := upload(ctx, key, f, fh.Size, fh.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}

		// Add photo to the request
		p, err := s.repository.AddRequestPhoto(ctx, request.ID, url, "")
		if err != nil {
			return nil, err
		}
		request.Photos = append(request.Photos, *p)
		added++
	}

	return request, nil
}

// GetViolationRequestByID gets a violation request by its ID
func (s *PokerService) GetViolationRequestByID(ctx context.Context, requestID string) (*model.ViolationRequest, error) {
	return s.repository.GetViolationRequestByID(ctx, requestID)
}
