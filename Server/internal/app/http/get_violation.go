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
	serviceGetViolation interface {
		GetViolationByID(ctx context.Context, id model.ViolationID) (*model.Violation, error)
	}

	GetViolationHandler struct {
		name     string
		service  serviceGetViolation
		uploader *objectstorage.Uploader
	}
)

func NewGetViolationHandler(name string, service serviceGetViolation, uploader *objectstorage.Uploader) *GetViolationHandler {
	return &GetViolationHandler{
		name:     name,
		service:  service,
		uploader: uploader,
	}
}

func (h *GetViolationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	// Try PathValue first (Go 1.22+), fallback to manual parsing
	var violationID model.ViolationID
	if id := r.PathValue("id"); id != "" {
		violationID = model.ViolationID(id)
	} else {
		// Fallback: extract from path manually
		path := strings.TrimPrefix(r.URL.Path, "/api/violations/")
		if path == "" || path == r.URL.Path {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation ID is required")
			return
		}

		// Remove any trailing slashes or query params
		if idx := strings.Index(path, "/"); idx != -1 {
			path = path[:idx]
		}
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}

		violationID = model.ViolationID(path)
	}

	if len(violationID) == 0 {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid violation ID format")
		return
	}

	violation, err := h.service.GetViolationByID(r.Context(), violationID)
	if err != nil {
		if errors.Is(err, model.ErrorNotFound) {
			uhttp.SendErrorResponse(w, http.StatusNotFound, "violation not found")
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

	// Request response structure
	type RequestResponse struct {
		ID              string          `json:"id"`
		Status          string          `json:"status"`
		CreatedByUserID model.UserID    `json:"created_by_user_id"`
		Comment         string          `json:"comment,omitempty"`
		CreatedAt       time.Time       `json:"created_at"`
		Photos          []PhotoResponse `json:"photos"`
	}

	// Transform requests with public URLs for photos
	publicRequests := make([]RequestResponse, 0, len(violation.Requests))
	var allPhotos []model.ViolationRequestPhoto

	for _, req := range violation.Requests {
		// Collect all photos for backward compatibility
		allPhotos = append(allPhotos, req.Photos...)

		// Transform photos for this request to public URLs
		requestPhotos := make([]PhotoResponse, 0, len(req.Photos))
		for _, photo := range req.Photos {
			publicPhoto := PhotoResponse{
				ID:           photo.ID,
				ViolationID:  string(violation.ID),
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

			requestPhotos = append(requestPhotos, publicPhoto)
		}

		publicRequests = append(publicRequests, RequestResponse{
			ID:              req.ID,
			Status:          string(req.Status),
			CreatedByUserID: req.CreatedByUserID,
			Comment:         req.Comment,
			CreatedAt:       req.CreatedAt,
			Photos:          requestPhotos,
		})
	}

	// Transform all photos to public URLs (for backward compatibility)
	publicPhotos := make([]PhotoResponse, 0, len(allPhotos))
	for _, photo := range allPhotos {
		publicPhoto := PhotoResponse{
			ID:           photo.ID,
			ViolationID:  string(violation.ID),
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

	// Return fields: user_id, description, lat, lng, photos (backward compatibility), requests
	response := struct {
		UserID      model.UserID      `json:"user_id"`
		Description string            `json:"description"`
		Lat         float64           `json:"lat"`
		Lng         float64           `json:"lng"`
		Photos      []PhotoResponse   `json:"photos"`
		Requests    []RequestResponse `json:"requests"`
	}{
		UserID:      violation.UserID,
		Description: violation.Description,
		Lat:         violation.Lat,
		Lng:         violation.Lng,
		Photos:      publicPhotos,
		Requests:    publicRequests,
	}

	b, err := json.Marshal(response)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uhttp.SendSuccessfulResponse(w, b)
}
