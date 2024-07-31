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
}

type Auth struct {
	bearerSecret   string
	bearerDuration int
}

func NewAuthorization(config *config.Configuration) (*Auth, error) {
	if config.Auth.BearerSecret == "" && config.Auth.BearerDuration == 0 {
		return nil, errors.New("auth components are missing")
	}

	return &Auth{bearerSecret: config.Auth.BearerSecret, bearerDuration: config.Auth.BearerDuration}, nil
}

type Claims struct {
	AccountNumber string `json:"accountNumber"`
	Role          string `json:"role"`
	Status        string `json:"status"`
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

func (a *Auth) Validate(tokenString string) (*jwt.Token, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.bearerSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, models.ErrorTokenSignature
		}
		return nil, err
	}

	if !token.Valid {
		return nil, models.ErrorInvalidToken
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, models.ErrorExpiredToken
	}

	return token, nil
}
