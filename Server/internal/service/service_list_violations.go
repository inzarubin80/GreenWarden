package service

import (
	"context"

	"github.com/inzarubin80/Server/internal/model"
)

func (s *PokerService) ListViolations(ctx context.Context, f *model.ListViolationsFilters) (*model.PaginatedViolations, error) {
	// Validate and clamp
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.MinLng != nil && (*f.MinLng < -180 || *f.MinLng > 180) {
		return nil, ErrBadRequest("invalid minLng")
	}
	if f.MaxLng != nil && (*f.MaxLng < -180 || *f.MaxLng > 180) {
		return nil, ErrBadRequest("invalid maxLng")
	}
	if f.MinLat != nil && (*f.MinLat < -90 || *f.MinLat > 90) {
		return nil, ErrBadRequest("invalid minLat")
	}
	if f.MaxLat != nil && (*f.MaxLat < -90 || *f.MaxLat > 90) {
		return nil, ErrBadRequest("invalid maxLat")
	}
	if f.From != nil && f.To != nil && f.From.After(*f.To) {
		// swap if needed
		tmp := *f.From
		*f.From = *f.To
		*f.To = tmp
	}

	items, total, err := s.repository.ListViolations(ctx, f)
	if err != nil {
		return nil, err
	}

	return &model.PaginatedViolations{
		Items:    items,
		Page:     f.Page,
		PageSize: f.PageSize,
		Total:    total,
	}, nil
}

// small error helper
type badRequest struct{ msg string }

func (e badRequest) Error() string { return e.msg }

func ErrBadRequest(msg string) error { return badRequest{msg: msg} }


