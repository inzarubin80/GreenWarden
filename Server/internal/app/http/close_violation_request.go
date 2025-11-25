package http

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/storage/objectstorage"
)

type (
	closeViolationRequestService interface {
		CreateViolationRequestWithPhotos(ctx context.Context, violationID model.ViolationID, status model.ViolationRequestStatus, userID model.UserID, comment string, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.ViolationRequest, error)
	}

	CloseViolationRequestHandler struct {
		name      string
		store     *sessions.CookieStore
		service   closeViolationRequestService
		uploader  *objectstorage.Uploader
		maxPhotos int
	}
)

func NewCloseViolationRequestHandler(store *sessions.CookieStore, name string, service closeViolationRequestService, uploader *objectstorage.Uploader, maxPhotos int) *CloseViolationRequestHandler {
	return &CloseViolationRequestHandler{
		name:      name,
		store:     store,
		service:   service,
		uploader:  uploader,
		maxPhotos: maxPhotos,
	}
}

func (h *CloseViolationRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form
	_ = r.ParseMultipartForm(32 << 20) // 32MB in-memory

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract violation ID from URL path
	var violationID model.ViolationID
	if id := r.PathValue("id"); id != "" {
		violationID = model.ViolationID(id)
	} else {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation ID is required")
		return
	}

	if len(violationID) == 0 {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid violation ID format")
		return
	}

	// Get status from form (required)
	statusStr := r.FormValue("status")
	if statusStr == "" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "status is required (partially_closed or closed)")
		return
	}

	status := model.ViolationRequestStatus(statusStr)
	if status != "partially_closed" && status != "closed" {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "status must be 'partially_closed' or 'closed'")
		return
	}

	// Get comment (optional)
	comment := r.FormValue("comment")

	// Get photos (optional)
	var files []*multipart.FileHeader
	if r.MultipartForm != nil {
		files = r.MultipartForm.File["photos"]
	}

	upload := func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error) {
		return h.uploader.Upload(ctx, key, file, size, contentType)
	}

	request, err := h.service.CreateViolationRequestWithPhotos(ctx, violationID, status, userID, comment, files, h.maxPhotos, upload)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := json.Marshal(request)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	uhttp.SendSuccessfulResponse(w, resp)
}

