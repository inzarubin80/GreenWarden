package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/storage/objectstorage"
)

type (
	serviceGetViolationRequest interface {
		GetViolationRequestByID(ctx context.Context, requestID string) (*model.ViolationRequest, error)
	}

	GetViolationRequestHandler struct {
		name     string
		service  serviceGetViolationRequest
		uploader *objectstorage.Uploader
	}
)

func NewGetViolationRequestHandler(name string, service serviceGetViolationRequest, uploader *objectstorage.Uploader) *GetViolationRequestHandler {
	return &GetViolationRequestHandler{
		name:     name,
		service:  service,
		uploader: uploader,
	}
}

func (h *GetViolationRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract request ID from URL path
	// Path format: /api/violations/{violation_id}/requests/{request_id}
	var requestID string
	if id := r.PathValue("request_id"); id != "" {
		requestID = id
	} else {
		// Fallback: extract from path manually
		// Try to extract from /api/violations/{violation_id}/requests/{request_id}
		path := strings.TrimPrefix(r.URL.Path, "/api/violations/")
		if path == "" || path == r.URL.Path {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, "request ID is required")
			return
		}

		// Remove violation_id and /requests/ prefix
		if idx := strings.Index(path, "/requests/"); idx != -1 {
			path = path[idx+len("/requests/"):]
		} else {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid path format")
			return
		}

		// Remove any trailing slashes or query params
		if idx := strings.Index(path, "/"); idx != -1 {
			path = path[:idx]
		}
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}

		requestID = path
	}

	if len(requestID) == 0 {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid request ID format")
		return
	}

	request, err := h.service.GetViolationRequestByID(r.Context(), requestID)
	if err != nil {
		if errors.Is(err, model.ErrorNotFound) {
			uhttp.SendErrorResponse(w, http.StatusNotFound, "violation request not found")
			return
		}
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Photo response structure for API compatibility
	type PhotoResponse struct {
		ID           string `json:"id"`
		ViolationID  string `json:"violation_id"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumb_url,omitempty"`
	}

	// Transform photos to public URLs
	publicPhotos := make([]PhotoResponse, 0, len(request.Photos))
	for _, photo := range request.Photos {
		publicPhoto := PhotoResponse{
			ID:           photo.ID,
			ViolationID:  string(request.ViolationID),
			URL:          photo.URL,
			ThumbnailURL: photo.ThumbnailURL,
		}

		// Generate public URL for main photo (24 hour expiry for presigned URLs)
		if publicURL, err := h.uploader.GetPublicURL(r.Context(), photo.URL, 24*time.Hour); err == nil {
			publicPhoto.URL = publicURL
		}

		// Generate public URL for thumbnail if present
		if photo.ThumbnailURL != "" {
			if thumbURL, err := h.uploader.GetPublicURL(r.Context(), photo.ThumbnailURL, 24*time.Hour); err == nil {
				publicPhoto.ThumbnailURL = thumbURL
			}
		}

		publicPhotos = append(publicPhotos, publicPhoto)
	}

	// Response structure
	response := struct {
		ID              string          `json:"id"`
		ViolationID     model.ViolationID `json:"violation_id"`
		Status          string          `json:"status"`
		CreatedByUserID model.UserID    `json:"created_by_user_id"`
		Comment         string          `json:"comment,omitempty"`
		CreatedAt       time.Time       `json:"created_at"`
		UpdatedAt       time.Time       `json:"updated_at"`
		Photos          []PhotoResponse `json:"photos"`
	}{
		ID:              request.ID,
		ViolationID:     request.ViolationID,
		Status:          string(request.Status),
		CreatedByUserID: request.CreatedByUserID,
		Comment:         request.Comment,
		CreatedAt:       request.CreatedAt,
		UpdatedAt:       request.UpdatedAt,
		Photos:          publicPhotos,
	}

	b, err := json.Marshal(response)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, b)
}

