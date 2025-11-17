package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	serviceListViolations interface {
		ListViolations(ctx context.Context, f *model.ListViolationsFilters) (*model.PaginatedViolations, error)
	}

	ListViolationsHandler struct {
		name    string
		service serviceListViolations
	}
)

func NewListViolationsHandler(name string, service serviceListViolations) *ListViolationsHandler {
	return &ListViolationsHandler{name: name, service: service}
}

func (h *ListViolationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var f model.ListViolationsFilters
	if t := strings.TrimSpace(q.Get("type")); t != "" {
		f.Type = &t
	}
	if s := strings.TrimSpace(q.Get("status")); s != "" {
		f.Status = &s
	}
	if v := strings.TrimSpace(q.Get("from")); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = &ts
		}
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = &ts
		}
	}
	if bbox := strings.TrimSpace(q.Get("bbox")); bbox != "" {
		parts := strings.Split(bbox, ",")
		if len(parts) == 4 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				f.MinLng = &v
			}
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				f.MinLat = &v
			}
			if v, err := strconv.ParseFloat(parts[2], 64); err == nil {
				f.MaxLng = &v
			}
			if v, err := strconv.ParseFloat(parts[3], 64); err == nil {
				f.MaxLat = &v
			}
		}
	}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Page = n
		}
	}
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.PageSize = n
		}
	}

	res, err := h.service.ListViolations(r.Context(), &f)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		uhttp.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	uhttp.SendSuccessfulResponse(w, b)
}
