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

// CreateViolationWithRequestPhotos creates violation and attaches uploaded photos to the open request
func (s *PokerService) CreateViolationWithRequestPhotos(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.Violation, error) {

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

// GetViolationChat returns paginated chat messages for a violation.
func (s *PokerService) GetViolationChat(ctx context.Context, violationID model.ViolationID, page, pageSize int) (*model.PaginatedViolationChatMessages, error) {
	items, total, err := s.repository.ListViolationChatMessages(ctx, violationID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Enrich messages with user display names.
	userIDSet := make(map[model.UserID]struct{})
	for _, m := range items {
		if m == nil {
			continue
		}
		if m.IsSystem {
			continue
		}
		if m.UserID == 0 {
			continue
		}
		userIDSet[m.UserID] = struct{}{}
	}
	if len(userIDSet) > 0 {
		userIDs := make([]model.UserID, 0, len(userIDSet))
		for id := range userIDSet {
			userIDs = append(userIDs, id)
		}
		users, err := s.repository.GetUsersByIDs(ctx, userIDs)
		if err == nil {
			nameByID := make(map[model.UserID]string, len(users))
			for _, u := range users {
				if u == nil {
					continue
				}
				nameByID[u.ID] = u.Name
			}
			for _, m := range items {
				if m == nil || m.IsSystem {
					continue
				}
				if name, ok := nameByID[m.UserID]; ok && name != "" {
					m.UserName = name
				}
			}
		}
	}

	return &model.PaginatedViolationChatMessages{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

// PostViolationChatMessage creates a new chat message and broadcasts it via Hub.
func (s *PokerService) PostViolationChatMessage(ctx context.Context, userID model.UserID, violationID model.ViolationID, text string, isSystem bool) (*model.ViolationChatMessage, error) {
	if text == "" {
		return nil, fmt.Errorf("message text is required")
	}

	// Ensure violation exists and user has access; reuse GetViolationByID for now.
	if _, err := s.repository.GetViolationByID(ctx, violationID); err != nil {
		return nil, err
	}

	msg, err := s.repository.CreateViolationChatMessage(ctx, violationID, userID, text, isSystem)
	if err != nil {
		return nil, err
	}

	// Enrich with user name for immediate WS payload.
	if !isSystem && userID != 0 {
		if u, err := s.repository.GetUser(ctx, userID); err == nil && u != nil {
			msg.UserName = u.Name
		}
	}

	// Broadcast to all subscribers for this violation as WS event.
	event := map[string]any{
		"type":    "message",
		"payload": msg,
	}
	_ = s.hub.BroadcastViolationMessage(violationID, event)

	return msg, nil
}

// UpdateViolationChatMessage updates text of a chat message if it belongs to the user.
func (s *PokerService) UpdateViolationChatMessage(ctx context.Context, userID model.UserID, messageID string, text string) (*model.ViolationChatMessage, error) {
	if text == "" {
		return nil, fmt.Errorf("message text is required")
	}
	msg, err := s.repository.UpdateViolationChatMessage(ctx, messageID, userID, text)
	if err != nil {
		return nil, err
	}

	// Enrich with user name for payload.
	if !msg.IsSystem && msg.UserID != 0 {
		if u, err := s.repository.GetUser(ctx, msg.UserID); err == nil && u != nil {
			msg.UserName = u.Name
		}
	}

	event := map[string]any{
		"type":    "message_updated",
		"payload": msg,
	}
	_ = s.hub.BroadcastViolationMessage(msg.ViolationID, event)

	return msg, nil
}

// DeleteViolationChatMessage deletes a chat message if it belongs to the user and notifies subscribers.
func (s *PokerService) DeleteViolationChatMessage(ctx context.Context, userID model.UserID, messageID string) error {
	vID, err := s.repository.DeleteViolationChatMessage(ctx, messageID, userID)
	if err != nil {
		return err
	}

	event := map[string]any{
		"type": "message_deleted",
		"payload": map[string]any{
			"id": messageID,
		},
	}
	_ = s.hub.BroadcastViolationMessage(vID, event)

	return nil
}
