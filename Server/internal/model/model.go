package model

import (
	"time"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	Access_Token_Type         = "access_token"
	Refresh_Token_Type        = "refresh_Token"
)

type (
	ViolationID string

	ViolationType   string
	ViolationStatus string

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
		ID                 UserID
		Name               string
		EvaluationStrategy string
		MaximumScore       int
	}

	Violation struct {
		ID                  ViolationID     `json:"id"`
		UserID              UserID          `json:"user_id"`
		Type                ViolationType   `json:"type"`
		Description         string          `json:"description"`
		Lat                 float64         `json:"lat"`
		Lng                 float64         `json:"lng"`
		Status              ViolationStatus `json:"status"`
		ConfirmationsCount  int             `json:"confirmations_count"`
		CreatedAt           time.Time       `json:"created_at"`
		UpdatedAt           time.Time       `json:"updated_at"`
	}

	LastSessionPoker struct {
		PokerID     PokerID
		UserID      UserID
		Name        string
	    IsAdmin     bool
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


	
)

func (p PokerID) UUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.MustParse(string(p)),
		Valid: true,
	}
}
