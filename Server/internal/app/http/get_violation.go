package http

import (
	"context"
	"encoding/json"
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
		if err == model.ErrorNotFound {
			uhttp.SendErrorResponse(w, http.StatusNotFound, "violation not found")
			return
		}
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Transform photo URLs to public URLs
	publicPhotos := make([]model.ViolationPhoto, len(violation.Photos))
	for i, photo := range violation.Photos {
		publicPhoto := photo
		
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
		
		publicPhotos[i] = publicPhoto
	}
	
	// Return only required fields: user_id, description, lat, lng, photos
	response := struct {
		UserID      model.UserID          `json:"user_id"`
		Description string                `json:"description"`
		Lat         float64               `json:"lat"`
		Lng         float64               `json:"lng"`
		Photos      []model.ViolationPhoto `json:"photos"`
	}{
		UserID:      violation.UserID,
		Description: violation.Description,
		Lat:         violation.Lat,
		Lng:         violation.Lng,
		Photos:      publicPhotos,
	}
	
	b, err := json.Marshal(response)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	uhttp.SendSuccessfulResponse(w, b)
}

