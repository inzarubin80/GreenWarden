package http

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/inzarubin80/Server/internal/app/uhttp"
	"github.com/inzarubin80/Server/internal/model"
	"github.com/inzarubin80/Server/internal/storage/objectstorage"
)

type (
	violationShareService interface {
		GetViolationByID(ctx context.Context, id model.ViolationID, userID model.UserID) (*model.Violation, error)
		GetViolationChat(ctx context.Context, violationID model.ViolationID, page, pageSize int) (*model.PaginatedViolationChatMessages, error)
	}

	ViolationSharePageHandler struct {
		name     string
		service  violationShareService
		uploader *objectstorage.Uploader
	}
)

func NewViolationSharePageHandler(name string, service violationShareService, uploader *objectstorage.Uploader) *ViolationSharePageHandler {
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

	ctx := r.Context()
	violation, err := h.service.GetViolationByID(ctx, model.ViolationID(id), 0)
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
	// Always open the public SPA for the map view (even if this SSR page is served via api.*).
	mapURL := "https://green-warden.ru/map/violation/" + id

	openRequest, resolutionRequest := splitRequests(violation.Requests)

	// og:title
	title := "GreenWarden — проблема на карте"
	if strings.TrimSpace(violation.Description) != "" {
		title = "GreenWarden — " + strings.TrimSpace(violation.Description)
	}

	// og:description
	description := buildDescription(violation, openRequest)

	// og:image: first photo of open request if present
	ogImage := firstPublicPhotoURL(ctx, h.uploader, openRequest)

	// Build public photo lists (gallery)
	beforePhotos := publicizeRequestPhotos(ctx, h.uploader, openRequest)
	afterPhotos := publicizeRequestPhotos(ctx, h.uploader, resolutionRequest)

	// Chat preview (last N messages)
	chatPreview, chatTotal := getChatPreview(ctx, h.service, model.ViolationID(id), 20)

	// Participants (from requests + chat)
	participants := buildParticipants(openRequest, resolutionRequest, chatPreview)

	// Analytics
	totalPhotos := len(beforePhotos) + len(afterPhotos)
	totalRequests := len(violation.Requests)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	vm := shareViewModel{
		Title:       title,
		Description: description,
		Canonical:   canonicalURL,
		MapURL:      mapURL,
		Image:       ogImage,

		ViolationID:          id,
		ViolationStatus:      string(violation.Status),
		TotalRequests:        totalRequests,
		TotalPhotos:          totalPhotos,
		ParticipantsCount:    len(participants),
		ChatMessagesTotal:    chatTotal,
		Coords:              formatCoords(violation.Lat, violation.Lng),
		FixedAt:             formatTime(openRequestCreatedAt(openRequest)),
		ResolvedAt:          formatTime(resolutionRequestCreatedAt(resolutionRequest)),
		OpenRequest:         requestCard(openRequest, "Зафиксировано", "Что заметили и почему это важно"),
		ResolutionRequest:   requestCard(resolutionRequest, resolutionTitle(resolutionRequest), "Что изменилось и как помогли участники"),
		BeforePhotos:        beforePhotos,
		AfterPhotos:         afterPhotos,
		Participants:        participants,
		ChatPreviewMessages: chatPreview,
	}

	_ = ogTemplate.Execute(w, vm)
}

func formatCoords(lat, lng float64) string {
	return fmt.Sprintf("%.5f, %.5f", lat, lng)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// Friendly format for humans
	return t.Local().Format("02.01.2006 15:04")
}

type sharePhoto struct {
	URL   string
	Thumb string
}

type shareParticipant struct {
	UserID        model.UserID
	Name          string
	AvatarURL     string
	BoostyURL     string
	Badges        []string
	PhotosCount   int
	MessagesCount int
}

type shareChatMessage struct {
	UserID      model.UserID
	UserName    string
	UserAvatar  string
	UserBoosty  string
	Text        string
	IsSystem    bool
	CreatedAt   string
}

type shareRequestCard struct {
	Title      string
	Subtitle   string
	AuthorName string
	AvatarURL  string
	BoostyURL  string
	Comment    string
	When       string
	Status     string
}

type shareViewModel struct {
	Title       string
	Description string
	Canonical   string
	MapURL      string
	Image       string

	ViolationID        string
	ViolationStatus    string
	Coords             string
	FixedAt            string
	ResolvedAt         string
	TotalRequests      int
	TotalPhotos        int
	ParticipantsCount  int
	ChatMessagesTotal  int64

	OpenRequest       shareRequestCard
	ResolutionRequest shareRequestCard

	BeforePhotos []sharePhoto
	AfterPhotos  []sharePhoto

	Participants        []shareParticipant
	ChatPreviewMessages []shareChatMessage
}

func splitRequests(reqs []model.ViolationRequest) (*model.ViolationRequest, *model.ViolationRequest) {
	var openReq *model.ViolationRequest
	var resolution *model.ViolationRequest

	for i := range reqs {
		r := reqs[i]
		if openReq == nil && r.Status == "open" {
			openReq = &r
		}
		if r.Status == "closed" || r.Status == "partially_closed" {
			// pick the latest by CreatedAt
			if resolution == nil || r.CreatedAt.After(resolution.CreatedAt) {
				tmp := r
				resolution = &tmp
			}
		}
	}
	if openReq == nil && len(reqs) > 0 {
		r := reqs[0]
		openReq = &r
	}
	return openReq, resolution
}

func buildDescription(v *model.Violation, openReq *model.ViolationRequest) string {
	descParts := []string{}
	if strings.TrimSpace(v.Description) != "" {
		descParts = append(descParts, strings.TrimSpace(v.Description))
	} else {
		descParts = append(descParts, "Проблема")
	}
	descParts = append(descParts, "Координаты: "+formatCoords(v.Lat, v.Lng))
	if openReq != nil && !openReq.CreatedAt.IsZero() {
		descParts = append(descParts, "Зафиксировано: "+openReq.CreatedAt.Format(time.RFC3339))
	}
	return strings.Join(descParts, " • ")
}

func firstPublicPhotoURL(ctx context.Context, uploader *objectstorage.Uploader, req *model.ViolationRequest) string {
	if req == nil || len(req.Photos) == 0 {
		return ""
	}
	photo := req.Photos[0]
	if strings.TrimSpace(photo.URL) == "" {
		return ""
	}
	out := photo.URL
	if uploader != nil {
		if publicURL, err := uploader.GetPublicURL(ctx, photo.URL, 24*time.Hour); err == nil {
			out = publicURL
		}
	}
	return out
}

func publicizeRequestPhotos(ctx context.Context, uploader *objectstorage.Uploader, req *model.ViolationRequest) []sharePhoto {
	if req == nil || len(req.Photos) == 0 {
		return nil
	}
	out := make([]sharePhoto, 0, len(req.Photos))
	for _, p := range req.Photos {
		if strings.TrimSpace(p.URL) == "" {
			continue
		}
		url := p.URL
		thumb := p.ThumbnailURL
		if uploader != nil {
			if publicURL, err := uploader.GetPublicURL(ctx, p.URL, 24*time.Hour); err == nil {
				url = publicURL
			}
			if p.ThumbnailURL != "" {
				if tURL, err := uploader.GetPublicURL(ctx, p.ThumbnailURL, 24*time.Hour); err == nil {
					thumb = tURL
				}
			}
		}
		if thumb == "" {
			thumb = url
		}
		out = append(out, sharePhoto{URL: url, Thumb: thumb})
	}
	return out
}

func openRequestCreatedAt(r *model.ViolationRequest) time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.CreatedAt
}

func resolutionRequestCreatedAt(r *model.ViolationRequest) time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.CreatedAt
}

func resolutionTitle(r *model.ViolationRequest) string {
	if r == nil {
		return "Не решено"
	}
	if r.Status == "partially_closed" {
		return "Частично решено"
	}
	if r.Status == "closed" {
		return "Решено"
	}
	return "Обновление"
}

func requestCard(r *model.ViolationRequest, title, subtitle string) shareRequestCard {
	if r == nil {
		return shareRequestCard{Title: title, Subtitle: subtitle}
	}
	name := strings.TrimSpace(r.AuthorName)
	if name == "" && r.CreatedByUserID != 0 {
		name = fmt.Sprintf("Пользователь #%d", r.CreatedByUserID)
	}
	return shareRequestCard{
		Title:      title,
		Subtitle:   subtitle,
		AuthorName: name,
		AvatarURL:  r.AuthorAvatarURL,
		BoostyURL:  r.AuthorBoostyURL,
		Comment:    strings.TrimSpace(r.Comment),
		When:       formatTime(r.CreatedAt),
		Status:     string(r.Status),
	}
}

func getChatPreview(ctx context.Context, svc violationShareService, violationID model.ViolationID, want int) ([]shareChatMessage, int64) {
	if want <= 0 {
		want = 20
	}

	// First call to get total
	first, err := svc.GetViolationChat(ctx, violationID, 1, 1)
	if err != nil || first == nil {
		return nil, 0
	}
	total := first.Total
	if total == 0 {
		return nil, 0
	}

	pageSize := 50
	lastPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	if lastPage <= 0 {
		lastPage = 1
	}
	last, err := svc.GetViolationChat(ctx, violationID, lastPage, pageSize)
	if err != nil || last == nil {
		return nil, total
	}

	// Take tail (messages are ASC)
	items := last.Items
	if len(items) > want {
		items = items[len(items)-want:]
	}

	out := make([]shareChatMessage, 0, len(items))
	for _, m := range items {
		if m == nil {
			continue
		}
		out = append(out, shareChatMessage{
			UserID:     m.UserID,
			UserName:   m.UserName,
			UserAvatar: m.UserAvatarURL,
			UserBoosty: m.UserBoostyURL,
			Text:       m.Text,
			IsSystem:   m.IsSystem,
			CreatedAt:  formatTime(m.CreatedAt),
		})
	}
	return out, total
}

func buildParticipants(openReq, resolutionReq *model.ViolationRequest, chat []shareChatMessage) []shareParticipant {
	type acc struct {
		p shareParticipant
	}
	byID := make(map[model.UserID]*shareParticipant)
	byName := make(map[string]*shareParticipant)

	ensure := func(userID model.UserID, name, avatar, boosty string) *shareParticipant {
		name = strings.TrimSpace(name)
		if userID != 0 {
			if p, ok := byID[userID]; ok {
				return p
			}
			p := &shareParticipant{UserID: userID, Name: name, AvatarURL: avatar, BoostyURL: boosty}
			byID[userID] = p
			return p
		}
		// fallback: key by name (for public/system messages)
		key := strings.ToLower(name)
		if key == "" {
			key = "unknown"
		}
		if p, ok := byName[key]; ok {
			return p
		}
		p := &shareParticipant{Name: name, AvatarURL: avatar, BoostyURL: boosty}
		byName[key] = p
		return p
	}

	addBadge := func(p *shareParticipant, badge string) {
		for _, b := range p.Badges {
			if b == badge {
				return
			}
		}
		p.Badges = append(p.Badges, badge)
	}

	if openReq != nil {
		p := ensure(openReq.CreatedByUserID, openReq.AuthorName, openReq.AuthorAvatarURL, openReq.AuthorBoostyURL)
		addBadge(p, "Зафиксировал")
		p.PhotosCount += len(openReq.Photos)
	}
	if resolutionReq != nil {
		p := ensure(resolutionReq.CreatedByUserID, resolutionReq.AuthorName, resolutionReq.AuthorAvatarURL, resolutionReq.AuthorBoostyURL)
		addBadge(p, "Решил")
		p.PhotosCount += len(resolutionReq.Photos)
	}

	for _, m := range chat {
		if m.IsSystem {
			continue
		}
		p := ensure(m.UserID, m.UserName, m.UserAvatar, m.UserBoosty)
		addBadge(p, "В чате")
		p.MessagesCount++
	}

	out := make([]shareParticipant, 0, len(byID)+len(byName))
	for _, p := range byID {
		out = append(out, *p)
	}
	for _, p := range byName {
		out = append(out, *p)
	}

	sort.Slice(out, func(i, j int) bool {
		// More “core” participants first: fixed/resolved, then counts.
		score := func(p shareParticipant) int {
			s := 0
			for _, b := range p.Badges {
				switch b {
				case "Решил":
					s += 1000
				case "Зафиксировал":
					s += 900
				case "В чате":
					s += 100
				}
			}
			s += p.PhotosCount*10 + p.MessagesCount
			return s
		}
		return score(out[i]) > score(out[j])
	})

	return out
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
      :root{
        --bg:#eaf7ef;
        --panel:rgba(255,255,255,.78);
        --border:rgba(15,23,42,.12);
        --text:#0f172a;
        --muted:rgba(15,23,42,.70);
        --muted2:rgba(15,23,42,.56);
        --brand:rgba(34,197,94,.24);
        --brand2:rgba(34,197,94,.14);
        --blue:rgba(59,130,246,.16);
      }
      body{margin:0;font-family:system-ui,-apple-system,Segoe UI,Roboto,Arial,sans-serif;background:radial-gradient(1200px 600px at 20% 10%, rgba(34,197,94,.20), transparent 60%),radial-gradient(900px 500px at 90% 20%, rgba(59,130,246,.12), transparent 55%),var(--bg);color:var(--text)}
      a{color:inherit}
      .wrap{max-width:980px;margin:0 auto;padding:20px}
      .shell{border:1px solid var(--border);background:var(--panel);backdrop-filter: blur(10px);-webkit-backdrop-filter: blur(10px);border-radius:20px;overflow:hidden;box-shadow:0 18px 70px rgba(15,23,42,.12)}
      .hero{padding:22px 22px 18px;border-bottom:1px solid var(--border)}
      .heroTop{display:flex;gap:12px;align-items:flex-start;justify-content:space-between;flex-wrap:wrap}
      .title{font-size:28px;line-height:1.15;margin:0}
      .sub{color:var(--muted);margin:10px 0 0;font-size:16px;line-height:1.5}
      .chips{display:flex;gap:8px;flex-wrap:wrap;margin-top:14px}
      .chip{display:inline-flex;align-items:center;gap:8px;border:1px solid var(--border);background:rgba(255,255,255,.65);padding:8px 10px;border-radius:999px;color:var(--muted);font-size:13px}
      .chip strong{color:var(--text);font-weight:650}
      .actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:16px}
      .btn{display:inline-flex;align-items:center;justify-content:center;gap:10px;padding:10px 14px;border-radius:999px;border:1px solid var(--border);background:rgba(255,255,255,.70);text-decoration:none}
      .btnPrimary{border-color:rgba(34,197,94,.35);background:linear-gradient(180deg, rgba(34,197,94,.20), rgba(34,197,94,.12))}
      .btnSecondary{border-color:rgba(59,130,246,.28);background:linear-gradient(180deg, rgba(59,130,246,.16), rgba(59,130,246,.08))}
      .btnSmall{padding:8px 10px;font-size:13px}
      .grid{display:grid;grid-template-columns:1.1fr .9fr;gap:16px;padding:18px 22px}
      @media (max-width: 920px){.grid{grid-template-columns:1fr;padding:16px}}
      .section{border:1px solid var(--border);background:rgba(255,255,255,.62);border-radius:16px;padding:16px}
      .sectionTitle{display:flex;align-items:baseline;justify-content:space-between;gap:12px;margin:0 0 10px}
      .sectionTitle h2{margin:0;font-size:18px}
      .sectionTitle p{margin:0;color:var(--muted2);font-size:13px}
      .timeline{display:grid;grid-template-columns:1fr 1fr;gap:12px}
      @media (max-width: 920px){.timeline{grid-template-columns:1fr}}
      .tcard{border:1px solid var(--border);background:rgba(255,255,255,.68);border-radius:14px;padding:14px}
      .tmeta{display:flex;gap:10px;align-items:center;margin-top:10px;color:var(--muted);font-size:13px}
      .avatar{width:34px;height:34px;border-radius:50%;border:1px solid var(--border);overflow:hidden;background:rgba(15,23,42,.06);display:flex;align-items:center;justify-content:center;font-weight:700}
      .avatar img{width:100%;height:100%;object-fit:cover}
      .badges{display:flex;gap:8px;flex-wrap:wrap;margin-top:8px}
      .badge{border:1px solid rgba(34,197,94,.28);background:rgba(34,197,94,.12);color:rgba(15,23,42,.82);padding:6px 10px;border-radius:999px;font-size:12px}
      .stats{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-top:14px}
      @media (max-width: 920px){.stats{grid-template-columns:repeat(2,1fr)}}
      .stat{border:1px solid var(--border);background:rgba(255,255,255,.70);border-radius:14px;padding:10px}
      .stat .k{font-size:12px;color:var(--muted2)}
      .stat .v{font-size:18px;font-weight:750;margin-top:4px}
      .gallery{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
      @media (max-width: 920px){.gallery{grid-template-columns:repeat(2,1fr)}}
      .gitem{border:1px solid var(--border);border-radius:14px;overflow:hidden;background:rgba(15,23,42,.03);cursor:pointer}
      .gitem img{width:100%;height:170px;object-fit:cover;display:block}
      .gcap{margin-top:10px;color:var(--muted2);font-size:13px}
      .participants{display:grid;grid-template-columns:repeat(2,1fr);gap:10px}
      @media (max-width: 920px){.participants{grid-template-columns:1fr}}
      .pcard{border:1px solid var(--border);background:rgba(255,255,255,.68);border-radius:14px;padding:12px;display:flex;gap:12px;align-items:flex-start}
      .pmeta{flex:1}
      .pname{font-weight:750;margin:0}
      .prow{display:flex;gap:8px;flex-wrap:wrap;margin-top:8px}
      .pdim{color:var(--muted2);font-size:12px}
      .donate{margin-top:10px;display:flex;gap:8px;flex-wrap:wrap}
      .chat{display:flex;flex-direction:column;gap:10px}
      .msg{border:1px solid var(--border);background:rgba(255,255,255,.68);border-radius:14px;padding:12px;display:flex;gap:10px;align-items:flex-start}
      .msgSystem{opacity:.72}
      .mhead{display:flex;gap:10px;align-items:center}
      .mname{font-weight:700}
      .mtime{color:var(--muted2);font-size:12px}
      .mtext{margin-top:6px;color:rgba(15,23,42,.86);line-height:1.45;white-space:pre-wrap}
      .footerNote{padding:14px 22px 22px;color:var(--muted2);font-size:12px}

      /* Lightbox */
      .lbOverlay{position:fixed;inset:0;background:rgba(15,23,42,.62);display:none;align-items:center;justify-content:center;padding:20px;z-index:9999}
      .lbOverlay.isOpen{display:flex}
      .lb{max-width:980px;width:100%;background:rgba(255,255,255,.94);border:1px solid rgba(255,255,255,.25);border-radius:16px;overflow:hidden}
      .lbTop{display:flex;align-items:center;justify-content:space-between;padding:10px 12px;border-bottom:1px solid rgba(15,23,42,.10)}
      .lbTop .ttl{font-weight:750}
      .lbBtns{display:flex;gap:8px}
      .lbImg{width:100%;max-height:72vh;object-fit:contain;background:#0b1220}
      .lbHint{padding:10px 12px;color:rgba(15,23,42,.62);font-size:12px}
    </style>
  </head>
  <body>
    <div class="wrap">
      <div class="shell">
        <div class="hero">
          <div class="heroTop">
            <div>
              <h1 class="title">{{ .Title }}</h1>
              <p class="sub">
                Страница о загрязнении и его решении: кто зафиксировал проблему на природе, какие действия были сделаны, и кто помог довести дело до результата.
              </p>
              <div class="chips">
                <div class="chip"><strong>Координаты</strong> {{ .Coords }}</div>
                {{- if .FixedAt }}<div class="chip"><strong>Зафиксировано</strong> {{ .FixedAt }}</div>{{- end }}
                {{- if .ResolvedAt }}<div class="chip"><strong>Решено</strong> {{ .ResolvedAt }}</div>{{- end }}
              </div>
              <div class="actions">
                <a class="btn btnPrimary" href="{{ .MapURL }}">Перейти на карту</a>
                <button class="btn btnSmall" type="button" id="copyLinkBtn">Скопировать ссылку</button>
              </div>
            </div>
          </div>

          <div class="stats">
            <div class="stat"><div class="k">Действий</div><div class="v">{{ .TotalRequests }}</div></div>
            <div class="stat"><div class="k">Фото</div><div class="v">{{ .TotalPhotos }}</div></div>
            <div class="stat"><div class="k">Участников</div><div class="v">{{ .ParticipantsCount }}</div></div>
            <div class="stat"><div class="k">Сообщений в чате</div><div class="v">{{ .ChatMessagesTotal }}</div></div>
          </div>
        </div>

        <div class="grid">
          <div>
            <div class="section">
              <div class="sectionTitle">
                <h2>Хронология</h2>
                <p>что произошло и кто помог природе</p>
              </div>
              <div class="timeline">
                <div class="tcard">
                  <div style="font-weight:750;font-size:16px">{{ .OpenRequest.Title }}</div>
                  <div style="color:var(--muted2);margin-top:6px">{{ .OpenRequest.Subtitle }}</div>
                  {{- if .OpenRequest.Comment }}<div class="mtext">{{ .OpenRequest.Comment }}</div>{{- end }}
                  <div class="tmeta">
                    <div class="avatar">
                      {{- if .OpenRequest.AvatarURL }}<img src="{{ .OpenRequest.AvatarURL }}" alt="avatar" />{{- else }}A{{- end }}
                    </div>
                    <div>
                      <div class="mname">{{ .OpenRequest.AuthorName }}</div>
                      {{- if .OpenRequest.When }}<div class="mtime">{{ .OpenRequest.When }}</div>{{- end }}
                    </div>
                    {{- if .OpenRequest.BoostyURL }}
                      <a class="btn btnSmall" href="{{ .OpenRequest.BoostyURL }}" target="_blank" rel="noreferrer">Поддержать</a>
                    {{- end }}
                  </div>
                </div>

                <div class="tcard">
                  <div style="font-weight:750;font-size:16px">{{ .ResolutionRequest.Title }}</div>
                  <div style="color:var(--muted2);margin-top:6px">{{ .ResolutionRequest.Subtitle }}</div>
                  {{- if .ResolutionRequest.Comment }}<div class="mtext">{{ .ResolutionRequest.Comment }}</div>{{- end }}
                  {{- if .ResolutionRequest.AuthorName }}
                  <div class="tmeta">
                    <div class="avatar">
                      {{- if .ResolutionRequest.AvatarURL }}<img src="{{ .ResolutionRequest.AvatarURL }}" alt="avatar" />{{- else }}A{{- end }}
                    </div>
                    <div>
                      <div class="mname">{{ .ResolutionRequest.AuthorName }}</div>
                      {{- if .ResolutionRequest.When }}<div class="mtime">{{ .ResolutionRequest.When }}</div>{{- end }}
                    </div>
                    {{- if .ResolutionRequest.BoostyURL }}
                      <a class="btn btnSmall" href="{{ .ResolutionRequest.BoostyURL }}" target="_blank" rel="noreferrer">Поддержать</a>
                    {{- end }}
                  </div>
                  {{- else }}
                  <div class="tmeta"><span class="mtime">Пока нет отметки о решении — вы можете помочь.</span></div>
                  {{- end }}
                </div>
              </div>
            </div>

            <div class="section" style="margin-top:16px">
              <div class="sectionTitle">
                <h2>Галерея</h2>
                <p>до / после — нажмите, чтобы увеличить</p>
              </div>
              {{- if .BeforePhotos }}
                <div class="gcap">Фото проблемы (до)</div>
                <div class="gallery" data-gallery="before">
                  {{- range $i, $p := .BeforePhotos }}
                  <div class="gitem" data-full="{{ $p.URL }}" data-idx="{{ $i }}" data-group="before">
                    <img src="{{ $p.Thumb }}" alt="Фото проблемы" loading="lazy" />
                  </div>
                  {{- end }}
                </div>
              {{- end }}
              {{- if .AfterPhotos }}
                <div class="gcap" style="margin-top:12px">Фото решения (после)</div>
                <div class="gallery" data-gallery="after">
                  {{- range $i, $p := .AfterPhotos }}
                  <div class="gitem" data-full="{{ $p.URL }}" data-idx="{{ $i }}" data-group="after">
                    <img src="{{ $p.Thumb }}" alt="Фото решения" loading="lazy" />
                  </div>
                  {{- end }}
                </div>
              {{- end }}
              {{- if not .BeforePhotos }}
                <div class="gcap">Фото пока нет — но вы уже можете помочь: поделиться ссылкой, поддержать участников, подключиться к уборке или фиксации.</div>
              {{- end }}
            </div>
          </div>

          <div>
            <div class="section">
              <div class="sectionTitle">
                <h2>Участники</h2>
                <p>люди, которые помогают природе</p>
              </div>
              <div class="gcap" style="margin-bottom:10px">
                Если вы хотите, чтобы уборок и решений было больше — поддержите участников. Донаты помогают покрывать дорогу, мешки, перчатки и вывоз.
              </div>
              <div class="participants">
                {{- range .Participants }}
                <div class="pcard">
                  <div class="avatar">
                    {{- if .AvatarURL }}<img src="{{ .AvatarURL }}" alt="avatar" />{{- else }}A{{- end }}
                  </div>
                  <div class="pmeta">
                    <p class="pname">{{ .Name }}</p>
                    <div class="badges">
                      {{- range .Badges }}<span class="badge">{{ . }}</span>{{- end }}
                    </div>
                    <div class="prow">
                      {{- if gt .PhotosCount 0 }}<span class="pdim">Фото: {{ .PhotosCount }}</span>{{- end }}
                      {{- if gt .MessagesCount 0 }}<span class="pdim">Сообщений: {{ .MessagesCount }}</span>{{- end }}
                    </div>
                    {{- if .BoostyURL }}
                    <div class="donate">
                      <a class="btn btnSmall btnPrimary" href="{{ .BoostyURL }}" target="_blank" rel="noreferrer">Поддержать донатом</a>
                    </div>
                    {{- end }}
                  </div>
                </div>
                {{- end }}
              </div>
            </div>

            <div class="section" style="margin-top:16px">
              <div class="sectionTitle">
                <h2>Чат</h2>
                <p>последние сообщения</p>
              </div>
              {{- if .ChatPreviewMessages }}
              <div class="chat">
                {{- range .ChatPreviewMessages }}
                <div class="msg {{ if .IsSystem }}msgSystem{{ end }}">
                  <div class="avatar">
                    {{- if .UserAvatar }}<img src="{{ .UserAvatar }}" alt="avatar" />{{- else }}A{{- end }}
                  </div>
                  <div style="flex:1">
                    <div class="mhead">
                      <div class="mname">{{ if .IsSystem }}Система{{ else }}{{ .UserName }}{{ end }}</div>
                      <div class="mtime">{{ .CreatedAt }}</div>
                      {{- if and (not .IsSystem) .UserBoosty }}
                        <a class="btn btnSmall" href="{{ .UserBoosty }}" target="_blank" rel="noreferrer">Донат</a>
                      {{- end }}
                    </div>
                    <div class="mtext">{{ .Text }}</div>
                  </div>
                </div>
                {{- end }}
              </div>
              {{- else }}
              <div class="gcap">В чате пока тихо. Обсуждение в приложении помогает собрать людей и довести уборку до результата.</div>
              {{- end }}
            </div>
          </div>
        </div>

        <div class="footerNote">
          Публичная страница для прозрачности: видно, кто зафиксировал загрязнение, какие действия были сделаны и кто помог. Делитесь ссылкой — так больше людей подключится к уборке и решению.
        </div>
      </div>
    </div>

    <!-- Lightbox -->
    <div class="lbOverlay" id="lbOverlay" role="dialog" aria-modal="true">
      <div class="lb" role="document">
        <div class="lbTop">
          <div class="ttl" id="lbTitle">Фото</div>
          <div class="lbBtns">
            <button type="button" class="btn btnSmall" id="lbPrev">←</button>
            <button type="button" class="btn btnSmall" id="lbNext">→</button>
            <button type="button" class="btn btnSmall" id="lbClose">Закрыть</button>
          </div>
        </div>
        <img class="lbImg" id="lbImg" src="" alt="Фото" />
        <div class="lbHint">Esc — закрыть • ←/→ — листать • клик по фону — закрыть</div>
      </div>
    </div>

    <script>
      (function() {
        // Copy link
        var copyBtn = document.getElementById('copyLinkBtn');
        if (copyBtn) {
          copyBtn.addEventListener('click', function() {
            try {
              navigator.clipboard.writeText({{ .Canonical | printf "%q" }});
              copyBtn.textContent = 'Скопировано';
              setTimeout(function(){ copyBtn.textContent = 'Скопировать ссылку'; }, 1600);
            } catch(e) {}
          });
        }

        // Lightbox
        var overlay = document.getElementById('lbOverlay');
        var img = document.getElementById('lbImg');
        var title = document.getElementById('lbTitle');
        var btnPrev = document.getElementById('lbPrev');
        var btnNext = document.getElementById('lbNext');
        var btnClose = document.getElementById('lbClose');

        var groups = { before: [], after: [] };
        var currentGroup = null;
        var currentIndex = 0;

        function collect() {
          var items = document.querySelectorAll('.gitem');
          items.forEach(function(el) {
            var g = el.getAttribute('data-group');
            var full = el.getAttribute('data-full');
            if (!g || !full) return;
            groups[g].push(full);
          });
        }
        function openLB(group, idx) {
          currentGroup = group;
          currentIndex = idx;
          var arr = groups[group] || [];
          if (!arr.length) return;
          if (currentIndex < 0) currentIndex = 0;
          if (currentIndex >= arr.length) currentIndex = arr.length - 1;
          img.src = arr[currentIndex];
          title.textContent = (group === 'after' ? 'После' : 'До') + ' — фото ' + (currentIndex + 1) + ' из ' + arr.length;
          overlay.classList.add('isOpen');
        }
        function closeLB() {
          overlay.classList.remove('isOpen');
          img.src = '';
        }
        function prev() { openLB(currentGroup, currentIndex - 1); }
        function next() { openLB(currentGroup, currentIndex + 1); }

        collect();
        document.querySelectorAll('.gitem').forEach(function(el) {
          el.addEventListener('click', function() {
            var g = el.getAttribute('data-group');
            var idx = parseInt(el.getAttribute('data-idx') || '0', 10);
            openLB(g, idx);
          });
        });

        btnPrev && btnPrev.addEventListener('click', prev);
        btnNext && btnNext.addEventListener('click', next);
        btnClose && btnClose.addEventListener('click', closeLB);
        overlay && overlay.addEventListener('click', function(e) {
          if (e.target === overlay) closeLB();
        });
        document.addEventListener('keydown', function(e) {
          if (!overlay.classList.contains('isOpen')) return;
          if (e.key === 'Escape') closeLB();
          if (e.key === 'ArrowLeft') prev();
          if (e.key === 'ArrowRight') next();
        });
      })();
    </script>
  </body>
</html>`))

