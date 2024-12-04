package authorization

import (
	"errors"
	"go-community/internal/config"
	"go-community/internal/models"
	"strings"
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
	bearerSecret    string
	bearerDuration  int
	refreshSecret   string
	refreshDuration int
}

func NewAuthorization(config *config.Configuration) (*Auth, error) {
	if config.Auth.BearerSecret == "" && config.Auth.BearerDuration == 0 && config.Auth.RefreshSecret == "" && config.Auth.RefreshDuration == 0 {
		return nil, errors.New("auth components are missing")
	}

	auth := &Auth{
		bearerSecret:    config.Auth.BearerSecret,
		bearerDuration:  config.Auth.BearerDuration,
		refreshSecret:   config.Auth.RefreshSecret,
		refreshDuration: config.Auth.RefreshDuration,
	}

	return auth, nil
}

type Claims struct {
	AccountNumber string `json:"accountNumber"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	jwt.RegisteredClaims
}

type Claim struct {
	CommunityId string   `json:"communityId"`
	UserTypes   []string `json:"userTypes"`
	Roles       []string `json:"roles"`
	Status      string   `json:"status"`
	jwt.RegisteredClaims
}

func (a *Auth) Generate(accountNumber string, role string, status string) (string, error) {
	duration := time.Now().Add(time.Duration(a.bearerDuration) * time.Minute)
	claims := &Claims{
		AccountNumber: accountNumber,
		Role:          strings.ToLower(role),
		Status:        strings.ToLower(status),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(duration),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.bearerSecret))
}

//func (a *Auth) Validate(tokenString string) (*jwt.Token, error) {
//	claims := &Claims{}
//	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
//		return []byte(a.bearerSecret), nil
//	})
//
//	if err != nil {
//		if errors.Is(err, jwt.ErrSignatureInvalid) {
//			return nil, models.ErrorTokenSignature
//		}
//		return nil, err
//	}
//
//	if !token.Valid {
//		return nil, models.ErrorInvalidToken
//	}
//
//	if claims.ExpiresAt.Time.Before(time.Now()) {
//		return nil, models.ErrorExpiredToken
//	}
//
//	return token, nil
//}

func (a *Auth) GenerateAccessToken(communityId string, userTypes []string, role []string, status string) (string, error) {
	expired := time.Now().Add(time.Duration(a.bearerDuration) * time.Minute)
	claims := &Claim{
		CommunityId: communityId,
		UserTypes:   userTypes,
		Roles:       role,
		Status:      strings.ToLower(status),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expired),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.bearerSecret))
}

func (a *Auth) GenerateRefreshToken(communityId string, userTypes []string, role []string, status string) (string, error) {
	expired := time.Now().Add(time.Duration(a.refreshDuration) * 24 * time.Hour)
	claims := &Claim{
		CommunityId: communityId,
		UserTypes:   userTypes,
		Roles:       role,
		Status:      strings.ToLower(status),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expired),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.refreshSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) GenerateTokens(communityId string, userTypes []string, role []string, status string) (*models.UserToken, error) {
	access, err := a.GenerateAccessToken(communityId, userTypes, role, status)
	if err != nil {
		return nil, err
	}

	refresh, err := a.GenerateRefreshToken(communityId, userTypes, role, status)
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
