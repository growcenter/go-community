package authorization

import (
	"encoding/base64"
	"errors"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authorization interface {
	Generate(accountNumber string) (string, error)
	Validate(token string) (*jwt.Token, error)
	GenerateAccessToken(communityId string, role string, status string) (string, error)
	GenerateRefreshToken(communityId string, role string, status string) (string, error)
	GenerateTokens(communityId string, role string, status string) (string, string, error)
}

type Auth struct {
	bearerKey       string
	bearerSecret    string
	bearerDuration  int
	refreshKey      string
	refreshSecret   string
	refreshDuration int
}

func NewAuthorization(config *config.Configuration) (*Auth, error) {
	if config.Auth.BearerSecret == nil && config.Auth.BearerDuration == 0 && config.Auth.RefreshSecret == nil && config.Auth.RefreshDuration == 0 {
		return nil, errors.New("auth components are missing")
	}

	var bearerKey, refreshKey string
	for bKey := range config.Auth.BearerSecret {
		bearerKey = bKey
	}

	for rKey := range config.Auth.RefreshSecret {
		refreshKey = rKey
	}

	auth := &Auth{
		bearerKey:       bearerKey,
		bearerSecret:    config.Auth.BearerSecret[bearerKey],
		bearerDuration:  config.Auth.BearerDuration,
		refreshKey:      refreshKey,
		refreshSecret:   config.Auth.RefreshSecret[refreshKey],
		refreshDuration: config.Auth.RefreshDuration,
	}

	return auth, nil
}

type Claim struct {
	Type string `json:"typ"`
	jwt.RegisteredClaims
	AuthorizedParty string   `json:"azp"`
	UserTypes       []string `json:"userTypes"`
	Roles           []string `json:"roles"`
}

func (a *Auth) GenerateAccessToken(id string, userTypes []string, role []string) (string, error) {
	now := common.Now()
	expired := now.Add(time.Duration(a.bearerDuration) * time.Minute)
	keyId := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(a.bearerKey))
	claims := &Claim{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			ExpiresAt: jwt.NewNumericDate(expired),
			IssuedAt:  jwt.NewNumericDate(now),
			Audience:  jwt.ClaimStrings{"otw"},
			Issuer:    "otw",
		},
		AuthorizedParty: "otw",
		Type:            "access",
		UserTypes:       userTypes,
		Roles:           role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = keyId
	tokenString, err := token.SignedString([]byte(a.bearerSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) GenerateRefreshToken(id string) (string, error) {
	now := common.Now()
	expired := now.Add(time.Duration(a.bearerDuration) * time.Minute)
	keyId := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(a.refreshKey))
	claims := &Claim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expired),
			Issuer:    "otw",
			Audience:  jwt.ClaimStrings{"otw"},
			Subject:   id,
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Type:            "refresh",
		AuthorizedParty: "otw",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = keyId
	tokenString, err := token.SignedString([]byte(a.refreshSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) GenerateTokens(id string, userTypes []string, role []string) (*models.UserToken, error) {
	access, err := a.GenerateAccessToken(id, userTypes, role)
	if err != nil {
		return nil, err
	}

	refresh, err := a.GenerateRefreshToken(id)
	if err != nil {
		return nil, err
	}

	tokens := models.UserToken{
		AccessToken:   access,
		AccessExpiry:  time.Now().Add(time.Duration(a.bearerDuration) * time.Minute),
		RefreshToken:  refresh,
		RefreshExpiry: time.Now().Add(time.Duration(a.refreshDuration) * 24 * time.Hour),
	}

	return &tokens, nil
}
