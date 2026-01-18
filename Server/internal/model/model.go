package model

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	Access_Token_Type  = "access_token"
	Refresh_Token_Type = "refresh_Token"
)

type (
	ViolationID string

	ViolationType          string
	ViolationStatus        string
	ViolationRequestStatus string

	TaskID     int64
	PokerID    string
	UserID     int64
	Estimate   int64
	CommentID  int64
	EstimateID int64

	UserProfileFromProvider struct {
		ProviderID   string `json:"provider_id"`   // Идентификатор пользователя у провайдера
		Email        string `json:"email"`         // Email пользователя
		Name         string `json:"name"`          // Имя пользователя
		FirstName    string `json:"first_name"`    // Имя
		LastName     string `json:"last_name"`     // Фамилия
		AvatarURL    string `json:"avatar_url"`    // Ссылка на аватар
		ProviderName string `json:"provider_name"` // Название провайдера (например, "google", "github")
	}

	User struct {
		ID                 UserID `json:"id"`
		Name               string `json:"name"`
		EvaluationStrategy string `json:"evaluation_strategy,omitempty"`
		MaximumScore       int    `json:"maximum_score,omitempty"`
		AvatarURL          string `json:"avatar_url,omitempty"`
		BoostyURL          string `json:"boosty_url,omitempty"`
	}

	Violation struct {
		ID                 ViolationID     `json:"id"`
		UserID             UserID          `json:"user_id"`
		Type               ViolationType   `json:"type"`
		Description        string          `json:"description"`
		Lat                float64         `json:"lat"`
		Lng                float64         `json:"lng"`
		Status             ViolationStatus `json:"status"`
		ConfirmationsCount int             `json:"confirmations_count"`
		CreatedAt          time.Time       `json:"created_at"`
		UpdatedAt          time.Time       `json:"updated_at"`
		// Фото загружаются через заявки (Requests)
		Requests []ViolationRequest `json:"requests,omitempty"`
	}

	ViolationRequest struct {
		ID              string                  `json:"id"`
		ViolationID     ViolationID             `json:"violation_id"`
		Status          ViolationRequestStatus  `json:"status"`
		CreatedByUserID UserID                  `json:"created_by_user_id"`
		AuthorName      string                  `json:"author_name,omitempty"`
		Comment         string                  `json:"comment,omitempty"`
		Photos          []ViolationRequestPhoto `json:"photos,omitempty"`
		CreatedAt       time.Time               `json:"created_at"`
		UpdatedAt       time.Time               `json:"updated_at"`
		Likes           int64                   `json:"likes"`
		Dislikes        int64                   `json:"dislikes"`
		UserVote        string                  `json:"user_vote"`
		AuthorBoostyURL string                  `json:"author_boosty_url,omitempty"`
		AuthorAvatarURL string                  `json:"author_avatar_url,omitempty"`
	}

	ViolationRequestPhoto struct {
		ID           string    `json:"id"`
		RequestID    string    `json:"request_id"`
		URL          string    `json:"url"`
		ThumbnailURL string    `json:"thumb_url,omitempty"`
		CreatedAt    time.Time `json:"created_at"`
	}

	// Deprecated: используйте ViolationRequestPhoto
	ViolationPhoto struct {
		ID           string `json:"id"`
		ViolationID  string `json:"violation_id"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumb_url,omitempty"`
	}

	UserSettings struct {
		UserID             UserID
		EvaluationStrategy string
		MaximumScore       int
	}

	UserAuthProviders struct {
		UserID      UserID
		ProviderUid string
		Provider    string
		Name        string
	}

	UserProfile struct {
		ID                 UserID   `json:"id"`
		Name               string   `json:"name"`
		AvatarURL          string   `json:"avatar_url,omitempty"`
		BoostyURL          string   `json:"boosty_url,omitempty"`
		ConnectedProviders []string `json:"connected_providers,omitempty"`
	}

	AuthData struct {
		UserID       UserID
		RefreshToken string
		AccessToken  string
	}

	Claims struct {
		UserID    UserID `json:"user_id"`
		TokenType string `json:"token_type"` // Добавляем поле для типа токена
		jwt.StandardClaims
	}

	// Violation votes
	ViolationVotes struct {
		ViolationID ViolationID `json:"violation_id"`
		Likes       int64       `json:"likes"`
		Dislikes    int64       `json:"dislikes"`
		// \"like\" | \"dislike\" | \"\" (если пользователь не голосовал или неавторизован)
		UserVote string `json:"user_vote"`
	}

	// Violation complaint (user report)
	ViolationComplaint struct {
		ID          string      `json:"id"`
		ViolationID ViolationID `json:"violation_id"`
		UserID      UserID      `json:"user_id"`
		RequestID   string      `json:"request_id,omitempty"`
		Reason      string      `json:"reason,omitempty"`
		Message     string      `json:"message,omitempty"`
		CreatedAt   time.Time   `json:"created_at"`
	}
)

// Filters and pagination DTOs
type ListViolationsFilters struct {
	Type     *string
	Status   *string
	From     *time.Time
	To       *time.Time
	MinLng   *float64
	MinLat   *float64
	MaxLng   *float64
	MaxLat   *float64
	Page     int
	PageSize int
}

type PaginatedViolations struct {
	Items    []*Violation `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
}

// Violation chat
type ViolationChatMessage struct {
	ID          string      `json:"id"`
	ViolationID ViolationID `json:"violation_id"`
	UserID      UserID      `json:"user_id"`
	UserName    string      `json:"user_name,omitempty"`
	UserAvatarURL string    `json:"user_avatar_url,omitempty"`
	UserBoostyURL  string   `json:"user_boosty_url,omitempty"`
	Text        string      `json:"text"`
	IsSystem    bool        `json:"is_system"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   *time.Time  `json:"updated_at,omitempty"`
}

type PaginatedViolationChatMessages struct {
	Items    []*ViolationChatMessage `json:"items"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Total    int64                   `json:"total"`
}

func (p PokerID) UUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.MustParse(string(p)),
		Valid: true,
	}
}
