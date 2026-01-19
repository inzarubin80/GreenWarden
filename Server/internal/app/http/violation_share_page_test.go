package http

import (
  "context"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"

  "github.com/inzarubin80/Server/internal/model"
)

type fakeViolationSvc struct{ v *model.Violation }

func (f fakeViolationSvc) GetViolationByID(ctx context.Context, id model.ViolationID, userID model.UserID) (*model.Violation, error) {
  return f.v, nil
}

func (f fakeViolationSvc) GetViolationChat(ctx context.Context, violationID model.ViolationID, page, pageSize int) (*model.PaginatedViolationChatMessages, error) {
  return &model.PaginatedViolationChatMessages{
    Items:    []*model.ViolationChatMessage{},
    Page:     page,
    PageSize: pageSize,
    Total:    0,
  }, nil
}

type fakeUploader struct{ url string }

func (u fakeUploader) GetPublicURL(ctx context.Context, storedURL string, expiry any) (string, error) { // not used
  return u.url, nil
}

func TestViolationSharePageHandler_OGTags(t *testing.T) {
  v := &model.Violation{
    ID:          "v1",
    Description: "Свалка у леса",
    Lat:         55.7558,
    Lng:         37.6173,
    Requests: []model.ViolationRequest{
      {
        Status: "open",
        Photos: []model.ViolationRequestPhoto{{URL: "https://cdn.example.com/p.jpg"}},
      },
    },
  }

  h := &ViolationSharePageHandler{name: "GET /violations/{id}", service: fakeViolationSvc{v: v}, uploader: nil}

  r := httptest.NewRequest(http.MethodGet, "http://example.com/violations/123", nil)
  // PathValue isn't available in httptest without routing; set URL path to trigger fallback parsing.
  r.URL.Path = "/violations/123"

  w := httptest.NewRecorder()
  h.ServeHTTP(w, r)

  if w.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", w.Code)
  }

  body := w.Body.String()
  for _, mustContain := range []string{
    "property=\"og:title\"",
    "property=\"og:description\"",
    "property=\"og:url\"",
    "property=\"og:image\"",
  } {
    if !strings.Contains(body, mustContain) {
      t.Fatalf("expected body to contain %q\nbody=%s", mustContain, body)
    }
  }
}
