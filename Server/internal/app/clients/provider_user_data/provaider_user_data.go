package provideruserdata

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/inzarubin80/Server/internal/model"
	"golang.org/x/oauth2"
)

type ProviderUserData struct {
	url         string
	oauthConfig *oauth2.Config
	provider    string
}

func NewProviderUserData(url string, oauthConfig *oauth2.Config, provider string) *ProviderUserData {
	return &ProviderUserData{
		url:         url,
		oauthConfig: oauthConfig,
		provider:    provider,
	}
}

func (p *ProviderUserData) GetUserData(ctx context.Context, authorizationCode string, codeVerifier string) (*model.UserProfileFromProvider, error) {

	var token *oauth2.Token
	var err error
	if codeVerifier != "" {
		token, err = p.oauthConfig.Exchange(ctx, authorizationCode, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	} else {
		token, err = p.oauthConfig.Exchange(ctx, authorizationCode)
	}
	if err != nil {
		return nil, err
	}

	client := p.oauthConfig.Client(ctx, token)
	response, err := client.Get(p.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var profile map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return nil, err
	}

	// Используем разные обработчики для разных провайдеров
	switch p.provider {
	case "yandex":
		return p.parseYandexProfile(profile)
	case "google":
		return p.parseGoogleProfile(profile)
	case "github":
		return p.parseGitHubProfile(profile)
	default:
		return p.parseDefaultProfile(profile)
	}
}

func (p *ProviderUserData) parseYandexProfile(profile map[string]interface{}) (*model.UserProfileFromProvider, error) {
	displayName, _ := profile["real_name"].(string)
	providerID, _ := profile["id"].(string)
	defaultEmail, _ := profile["default_email"].(string)
	firstName, _ := profile["first_name"].(string)
	lastName, _ := profile["last_name"].(string)

	userData := &model.UserProfileFromProvider{
		Name:         displayName,
		ProviderID:   providerID,
		ProviderName: p.provider,
		Email:        defaultEmail,
		FirstName:    firstName,
		LastName:     lastName,
	}

	return userData, nil
}

func (p *ProviderUserData) parseGoogleProfile(profile map[string]interface{}) (*model.UserProfileFromProvider, error) {
	// Google API возвращает данные в другом формате
	displayName, _ := profile["name"].(string)
	providerID, _ := profile["id"].(string)
	email, _ := profile["email"].(string)
	firstName, _ := profile["given_name"].(string)
	lastName, _ := profile["family_name"].(string)

	// Если displayName пустой, используем комбинацию имени и фамилии
	if displayName == "" {
		if firstName != "" && lastName != "" {
			displayName = fmt.Sprintf("%s %s", firstName, lastName)
		} else if firstName != "" {
			displayName = firstName
		} else if lastName != "" {
			displayName = lastName
		}
	}

	userData := &model.UserProfileFromProvider{
		Name:         displayName,
		ProviderID:   providerID,
		ProviderName: p.provider,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
	}

	return userData, nil
}

func (p *ProviderUserData) parseGitHubProfile(profile map[string]interface{}) (*model.UserProfileFromProvider, error) {
	// GitHub API возвращает данные в своем формате
	displayName, _ := profile["name"].(string)
	providerID, _ := profile["id"].(float64) // GitHub возвращает ID как число
	email, _ := profile["email"].(string)
	login, _ := profile["login"].(string)

	// Если displayName пустой, используем login
	if displayName == "" {
		displayName = login
	}

	userData := &model.UserProfileFromProvider{
		Name:         displayName,
		ProviderID:   fmt.Sprintf("%.0f", providerID), // Конвертируем в строку
		ProviderName: p.provider,
		Email:        email,
		FirstName:    displayName, // GitHub не предоставляет отдельно имя и фамилию
		LastName:     "",
	}

	return userData, nil
}

func (p *ProviderUserData) parseDefaultProfile(profile map[string]interface{}) (*model.UserProfileFromProvider, error) {
	// Универсальный обработчик для неизвестных провайдеров
	displayName, _ := profile["name"].(string)
	providerID, _ := profile["id"].(string)
	email, _ := profile["email"].(string)
	firstName, _ := profile["first_name"].(string)
	lastName, _ := profile["last_name"].(string)

	userData := &model.UserProfileFromProvider{
		Name:         displayName,
		ProviderID:   providerID,
		ProviderName: p.provider,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
	}

	return userData, nil
}
