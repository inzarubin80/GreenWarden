package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/inzarubin80/Server/internal/app/defenitions"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/storage/objectstorage"
)

type (
	updateAvatarService interface {
		UpdateUserAvatar(ctx context.Context, userID model.UserID, avatarURL string) error
	}

	UpdateAvatarHandler struct {
		name     string
		service  updateAvatarService
		uploader *objectstorage.Uploader
	}
)

func NewUpdateAvatarHandler(name string, service updateAvatarService, uploader *objectstorage.Uploader) *UpdateAvatarHandler {
	return &UpdateAvatarHandler{
		name:     name,
		service:  service,
		uploader: uploader,
	}
}

func (h *UpdateAvatarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(defenitions.UserID).(model.UserID)
	if !ok {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "user ID not found in context")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file into memory to know size
	data, err := io.ReadAll(file)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, "failed to read avatar")
		return
	}

	key := fmt.Sprintf("avatars/%d/%d%s", userID, time.Now().UnixNano(), filepath.Ext(header.Filename))
	reader := bytes.NewReader(data)

	url, err := h.uploader.Upload(ctx, key, reader, int64(len(data)), contentType)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	if err := h.service.UpdateUserAvatar(ctx, userID, url); err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]string{
		"avatar_url": url,
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, "failed to marshal response")
		return
	}
	uhttp.SendSuccessfulResponse(w, jsonData)
}


