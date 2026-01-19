package http

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/storage/objectstorage"
)

type (
	ViolationSharePageHandler struct {
		name     string
		service  serviceGetViolation
		uploader *objectstorage.Uploader
	}
)

func NewViolationSharePageHandler(name string, service serviceGetViolation, uploader *objectstorage.Uploader) *ViolationSharePageHandler {
	return &ViolationSharePageHandler{name: name, service: service, uploader: uploader}
}

func (h *ViolationSharePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		// Fallback for older routers
		path := strings.TrimPrefix(r.URL.Path, "/violations/")
		if path == "" || path == r.URL.Path {
			uhttp.SendErrorResponse(w, http.StatusBadRequest, "violation id is required")
			return
		}
		if idx := strings.Index(path, "/"); idx != -1 {
			path = path[:idx]
		}
		id = path
	}

	violation, err := h.service.GetViolationByID(r.Context(), model.ViolationID(id), 0)
	if err != nil {
		if errors.Is(err, model.ErrorNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shareBase := os.Getenv("PUBLIC_WEB_BASE")
	if shareBase == "" {
		// fallback: infer from request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
			scheme = v
		}
		host := r.Host
		if v := r.Header.Get("X-Forwarded-Host"); v != "" {
			host = v
		}
		shareBase = scheme + "://" + host
	}

	canonicalURL := strings.TrimRight(shareBase, "/") + "/violations/" + id

	// Find open request (report) for a better description/photo source.
	var (
		openRequest *model.ViolationRequest
	)
	for i := range violation.Requests {
		req := violation.Requests[i]
		if req.Status == "open" {
			openRequest = &req
			break
		}
	}
	if openRequest == nil && len(violation.Requests) > 0 {
		openRequest = &violation.Requests[0]
	}

	// og:title
	title := "GreenWarden — проблема на карте"
	if strings.TrimSpace(violation.Description) != "" {
		title = "GreenWarden — " + strings.TrimSpace(violation.Description)
	}

	// og:description
	descParts := []string{}
	if strings.TrimSpace(violation.Description) != "" {
		descParts = append(descParts, strings.TrimSpace(violation.Description))
	}
	descParts = append(descParts, "Координаты: "+formatCoords(violation.Lat, violation.Lng))
	if openRequest != nil && !openRequest.CreatedAt.IsZero() {
		descParts = append(descParts, "Зафиксировано: "+openRequest.CreatedAt.Format(time.RFC3339))
	}
	description := strings.Join(descParts, " • ")

	// og:image: first photo of open request if present
	ogImage := ""
	if openRequest != nil && len(openRequest.Photos) > 0 {
		photo := openRequest.Photos[0]
		if strings.TrimSpace(photo.URL) != "" {
			// photo.URL is already a stored/public URL in many cases; set it as fallback.
			ogImage = photo.URL
			// If uploader is available, prefer CDN/presigned public URL.
			if h.uploader != nil {
				if publicURL, err := h.uploader.GetPublicURL(r.Context(), photo.URL, 24*time.Hour); err == nil {
					ogImage = publicURL
				}
			}
		}
	}

	// Render HTML with OG tags. Bots will parse <head>; humans can click through.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_ = ogTemplate.Execute(w, map[string]any{
		"Title":       title,
		"Description": description,
		"Canonical":   canonicalURL,
		"Image":       ogImage,
	})
}

func formatCoords(lat, lng float64) string {
	return fmt.Sprintf("%.5f, %.5f", lat, lng)
}

var ogTemplate = template.Must(template.New("og").Parse(`<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{ .Title }}</title>
    <meta name="description" content="{{ .Description }}" />

    <meta property="og:type" content="article" />
    <meta property="og:site_name" content="GreenWarden" />
    <meta property="og:title" content="{{ .Title }}" />
    <meta property="og:description" content="{{ .Description }}" />
    <meta property="og:url" content="{{ .Canonical }}" />
    {{- if .Image }}
    <meta property="og:image" content="{{ .Image }}" />
    {{- end }}

    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="{{ .Title }}" />
    <meta name="twitter:description" content="{{ .Description }}" />
    {{- if .Image }}
    <meta name="twitter:image" content="{{ .Image }}" />
    {{- end }}

    <link rel="canonical" href="{{ .Canonical }}" />
    <style>
      body{margin:0;font-family:system-ui,-apple-system,Segoe UI,Roboto,Arial,sans-serif;background:#eaf7ef;color:#0f172a}
      .wrap{max-width:880px;margin:0 auto;padding:24px}
      .card{border:1px solid rgba(15,23,42,.12);background:rgba(255,255,255,.72);border-radius:16px;padding:18px}
      .title{font-size:22px;margin:0 0 8px}
      .desc{color:rgba(15,23,42,.72);margin:0 0 16px;line-height:1.5}
      .btn{display:inline-flex;align-items:center;justify-content:center;padding:10px 14px;border-radius:999px;border:1px solid rgba(59,130,246,.35);background:rgba(37,99,235,.1);color:rgba(15,23,42,.9);text-decoration:none}
      .img{margin-top:14px;border-radius:14px;border:1px solid rgba(15,23,42,.12);overflow:hidden}
      .img img{width:100%;display:block}
      .note{margin-top:12px;font-size:12px;color:rgba(15,23,42,.6)}
    </style>
  </head>
  <body>
    <div class="wrap">
      <div class="card">
        <h1 class="title">{{ .Title }}</h1>
        <p class="desc">{{ .Description }}</p>
        <a class="btn" href="{{ .Canonical }}">Открыть проблему</a>
        {{- if .Image }}
        <div class="img"><img src="{{ .Image }}" alt="Фото проблемы"></div>
        {{- end }}
        <div class="note">Страница нужна для превью в соцсетях (OG/VK). Если вы человек — нажмите «Открыть проблему».</div>
      </div>
    </div>
  </body>
</html>`))

