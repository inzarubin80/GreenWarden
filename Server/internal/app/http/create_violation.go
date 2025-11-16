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
	createViolationService interface {
		CreateViolation(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64) (*model.Violation, error)
		CreateViolationWithPhotos(ctx context.Context, userID model.UserID, vType model.ViolationType, description string, lat, lng float64, files []*multipart.FileHeader, maxPhotos int, upload func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error)) (*model.Violation, error)
	}

	CreateViolationHandler struct {
		name      string
		store     *sessions.CookieStore
		service   createViolationService
		uploader  *objectstorage.Uploader
		maxPhotos int
	}
)

func NewCreateViolationHandler(store *sessions.CookieStore, name string, service createViolationService, uploader *objectstorage.Uploader, maxPhotos int) *CreateViolationHandler {
	return &CreateViolationHandler{
		name:      name,
		store:     store,
		service:   service,
		uploader:  uploader,
		maxPhotos: maxPhotos,
	}
}

func (h *CreateViolationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Try multipart first (for photos[]). If not multipart, fallback to JSON body.
	_ = r.ParseMultipartForm(32 << 20) // 32MB in-memory

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.MultipartForm != nil {
		vType := r.FormValue("type")
		description := r.FormValue("description")
		lat, lng := parseFloat(r.FormValue("lat")), parseFloat(r.FormValue("lng"))
		files := r.MultipartForm.File["photos"]
		upload := func(ctx context.Context, key string, file multipart.File, size int64, contentType string) (string, error) {
			return h.uploader.Upload(ctx, key, file, size, contentType)
		}
		v, err := h.service.CreateViolationWithPhotos(ctx, userID, model.ViolationType(vType), description, lat, lng, files, h.maxPhotos, upload)
		if err != nil {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, v)
		return
	}

	var req struct {
		Type        string  `json:"type"`
		Description string  `json:"description"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	v, err := h.service.CreateViolation(ctx, userID, model.ViolationType(req.Type), req.Description, req.Lat, req.Lng)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, v)
}

func jsonResponse(w http.ResponseWriter, v any) {
	resp, err := json.Marshal(v)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	uhttp.SendSuccessfulResponse(w, resp)
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	_ = json.Unmarshal([]byte(s), &f)
	return f
}
