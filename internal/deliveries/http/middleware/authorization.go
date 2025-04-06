package middleware

import (
	"encoding/base64"
	"github.com/google/uuid"
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/usecases"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type jwtClaim struct {
	UserTypes       []string `json:"userTypes"`
	Roles           []string `json:"roles"`
	Type            string   `json:"typ"`
	AuthorizedParty string   `json:"azp"`
	jwt.RegisteredClaims
}

func UserMiddleware(config *config.Configuration, usecase *usecases.Usecases, allowedRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header := ctx.Request().Header.Get("Authorization")
			if header == "" {
				return response.Error(ctx, models.ErrorEmptyToken)
			}

			tokenString := header[len("Bearer "):]
			token, err := jwt.ParseWithClaims(tokenString, &jwtClaim{}, func(token *jwt.Token) (sec interface{}, err error) {
				// Extract the Key ID from the token header
				if kid, ok := token.Header["kid"].(string); ok {
					keyid, err := base64.RawURLEncoding.DecodeString(kid)
					if err != nil {
						return nil, err
					}

					if key, exists := config.Auth.BearerSecret[string(keyid)]; exists {
						// This marks success
						return []byte(key), nil
					}
				}

				// This marks error
				return nil, err
			})

			if err != nil {
				if err.Error() == "token has invalid claims: token is expired" {
					return response.Error(ctx, models.ErrorExpiredToken)
				}
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			claims, ok := token.Claims.(*jwtClaim)
			if !ok || !token.Valid {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if claims.ExpiresAt.Time.Before(time.Now()) {
				return response.Error(ctx, models.ErrorExpiredToken)
			}

			if claims.IssuedAt.Time.After(time.Now()) {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if claims.Type != "access" {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if allowedRoles != nil {
				for _, userType := range claims.UserTypes {
					if userType == "superadmin" {
						ctx.Set("id", claims.Subject)
						ctx.Set("userTypes", claims.UserTypes)
						ctx.Set("roles", claims.Roles)
						return next(ctx)
					}
				}

				// Check if the user's roles match the required roles
				for _, allowedRole := range allowedRoles {
					for _, userRole := range claims.Roles {
						if userRole == allowedRole {
							ctx.Set("id", claims.Subject)
							ctx.Set("userTypes", claims.UserTypes)
							ctx.Set("roles", claims.Roles)
							return next(ctx)
						}
					}
				}

				return response.Error(ctx, models.ErrorForbiddenRole)
			}

			ctx.Set("id", claims.Subject)
			ctx.Set("userTypes", claims.UserTypes)
			ctx.Set("roles", claims.Roles)

			return next(ctx)
		}
	}
}

func RefreshMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header, err := ctx.Cookie("refresh_token")
			if err == nil && header.Value != "" {
				token, err := jwt.ParseWithClaims(header.Value, &jwtClaim{}, func(token *jwt.Token) (sec interface{}, err error) {
					if kid, ok := token.Header["kid"].(string); ok {
						keyid, err := base64.RawURLEncoding.DecodeString(kid)
						if err != nil {
							return nil, err
						}

						if key, exists := config.Auth.RefreshSecret[string(keyid)]; exists {
							// This marks success
							return []byte(key), nil
						}
					}

					// This marks error
					return nil, err
				})

				if err != nil {
					if err.Error() == "token has invalid claims: token is expired" {
						return response.Error(ctx, models.ErrorExpiredToken)
					}
					return response.Error(ctx, models.ErrorInvalidToken)
				}

				claims, ok := token.Claims.(*jwtClaim)
				if !ok || !token.Valid {
					return response.Error(ctx, models.ErrorInvalidToken)
				}

				if claims.ExpiresAt.Time.Before(time.Now()) {
					return response.Error(ctx, models.ErrorExpiredToken)
				}

				if claims.IssuedAt.Time.After(time.Now()) {
					return response.Error(ctx, models.ErrorInvalidToken)
				}

				if claims.Type != "refresh" {
					return response.Error(ctx, models.ErrorInvalidToken)
				}

				ctx.Set("id", claims.Subject)
				return next(ctx)
			}

			ctx.Set("id", uuid.New().String())
			return next(ctx)
		}
	}
}

func GeneralMiddleware(config *config.Configuration, usecase *usecases.Usecases) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			enableGeneralHeader, err := usecase.FeatureFlag.IsFeatureEnabled(ctx.Request().Context(), "event_be_enablegeneralheader", "")
			if err != nil {
				return response.Error(ctx, err)
			}

			if !enableGeneralHeader {
				key := ctx.Request().Header.Get("X-API-Key")
				if key == "" {
					return response.Error(ctx, models.ErrorEmptyAPIKey)
				}

				if key != config.Auth.APIKey {
					return response.Error(ctx, models.ErrorInvalidAPIKey)
				}

				return next(ctx)
			}

			// Extract or generate X-Request-Id
			requestId := ctx.Request().Header.Get("X-Request-Id")
			if requestId == "" {
				requestId = uuid.New().String() // Generate a new UUID if missing
			}

			timestamp := ctx.Request().Header.Get("X-Timestamp")
			if timestamp == "" {
				timestamp = common.Now().Format(time.RFC3339)
			}

			clientId := ctx.Request().Header.Get("X-Client-Id")
			if clientId != "" {
				decoded, err := base64.StdEncoding.DecodeString(clientId)
				if err != nil {
					return response.Error(ctx, err)
				}

				if config.Auth.ClientId[string(decoded)] == false {
					return response.Error(ctx, models.ErrorInvalidToken)
				}
			}

			origin := ctx.Request().Header.Get("Origin")
			if origin == "" {
				origin = "ongoing"
			}

			apiKey := ctx.Request().Header.Get("X-API-Key")
			if apiKey != "" && config.Auth.APIKey != apiKey {
				return response.Error(ctx, models.ErrorInvalidAPIKey)
			}

			// Store in context for later use
			ctx.Set("X-Request-Id", requestId)
			ctx.Set("X-Timestamp", timestamp)

			ctx.Response().Header().Set("X-Request-Id", requestId)
			ctx.Response().Header().Set("X-Timestamp", timestamp)

			return next(ctx) // Continue to the next middleware
		}
	}
}
