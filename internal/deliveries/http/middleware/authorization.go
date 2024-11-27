package middleware

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type jwtClaims struct {
	AccountNumber string `json:"accountNumber"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	jwt.RegisteredClaims
}

type jwtClaim struct {
	CommunityId string   `json:"communityId"`
	UserType    string   `json:"userType"`
	Roles       []string `json:"roles"`
	Status      string   `json:"status"`
	jwt.RegisteredClaims
}

func UserMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header := ctx.Request().Header.Get("Authorization")
			if header == "" {
				return response.Error(ctx, models.ErrorEmptyToken)
			}

			tokenString := header[len("Bearer "):]
			token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (sec interface{}, err error) {
				if config.Auth.BearerSecret == "" {
					return nil, err
				}
				return []byte(config.Auth.BearerSecret), nil
			})

			if err != nil {
				if err.Error() == "token has invalid claims: token is expired" {
					return response.Error(ctx, models.ErrorExpiredToken)
				}
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			claims, ok := token.Claims.(*jwtClaims)
			if !ok || !token.Valid {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if claims.ExpiresAt.Time.Before(time.Now()) {
				return response.Error(ctx, models.ErrorExpiredToken)
			}

			if strings.ToLower(claims.Status) != "active" || strings.ToLower(claims.Status) == "inactive" {
				return response.Error(ctx, models.ErrorLoggedOut)
			}

			ctx.Set("accountNumber", claims.AccountNumber)
			return next(ctx)
		}
	}
}

func InternalMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			key := ctx.Request().Header.Get("X-API-Key")
			if key == "" {
				return response.Error(ctx, models.ErrorEmptyAPIKey)
			}

			if key != config.Auth.APIKey {
				return response.Error(ctx, models.ErrorInvalidAPIKey)
			}

			return next(ctx)
		}
	}
}

func AdminMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header := ctx.Request().Header.Get("Authorization")
			if header == "" {
				return response.Error(ctx, models.ErrorEmptyToken)
			}

			tokenString := header[len("Bearer "):]
			token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (sec interface{}, err error) {
				if config.Auth.BearerSecret == "" {
					return nil, err
				}
				return []byte(config.Auth.BearerSecret), nil
			})

			if err != nil {
				if err.Error() == "token has invalid claims: token is expired" {
					return response.Error(ctx, models.ErrorExpiredToken)
				}
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			claims, ok := token.Claims.(*jwtClaims)
			if !ok || !token.Valid {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if claims.ExpiresAt.Time.Before(time.Now()) {
				return response.Error(ctx, models.ErrorExpiredToken)
			}

			if strings.ToLower(claims.Role) != "admin" && strings.ToLower(claims.Role) != "usher" {
				return response.Error(ctx, models.ErrorForbiddenRole)
			}

			if strings.ToLower(claims.Status) != "active" || strings.ToLower(claims.Status) == "inactive" {
				return response.Error(ctx, models.ErrorLoggedOut)
			}

			ctx.Set("accountNumber", claims.AccountNumber)
			return next(ctx)
		}
	}
}

func RoleUserMiddleware(config *config.Configuration, allowedRoles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header := ctx.Request().Header.Get("Authorization")
			if header == "" {
				return response.Error(ctx, models.ErrorEmptyToken)
			}

			tokenString := header[len("Bearer "):]
			token, err := jwt.ParseWithClaims(tokenString, &jwtClaim{}, func(token *jwt.Token) (sec interface{}, err error) {
				if config.Auth.BearerSecret == "" {
					return nil, err
				}
				return []byte(config.Auth.BearerSecret), nil
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

			if strings.ToLower(claims.Status) != "active" || strings.ToLower(claims.Status) == "inactive" {
				return response.Error(ctx, models.ErrorLoggedOut)
			}

			if strings.ToLower(claims.UserType) == "superadmin" {
				ctx.Set("communityId", claims.CommunityId)
				ctx.Set("userType", claims.UserType)
				ctx.Set("roles", claims.Roles)
				return next(ctx)
			}

			// Check if the user's roles match the required roles
			for _, allowedRole := range allowedRoles {
				for _, userRole := range claims.Roles {
					if userRole == allowedRole {
						ctx.Set("communityId", claims.CommunityId)
						ctx.Set("userType", claims.UserType)
						ctx.Set("roles", claims.Roles)
						return next(ctx)
					}
				}
			}

			return response.Error(ctx, models.ErrorForbiddenRole)
		}
	}
}

func RefreshMiddleware(config *config.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header, err := ctx.Cookie("refresh_token")

			if err != nil {
				response.Error(ctx, err)
			}

			if header.Value == "" {
				return response.Error(ctx, models.ErrorEmptyToken)
			}

			//tokenString := header[len("Bearer "):]
			token, err := jwt.ParseWithClaims(header.Value, &jwtClaim{}, func(token *jwt.Token) (sec interface{}, err error) {
				if config.Auth.RefreshSecret == "" {
					return nil, err
				}
				return []byte(config.Auth.RefreshSecret), nil
			})

			if err != nil {
				if err.Error() == "token has invalid claims: token is expired" {
					return response.Error(ctx, models.ErrorExpiredToken)
				}
				//return response.Error(ctx, models.ErrorInvalidToken)
				return response.Error(ctx, err)
			}

			claims, ok := token.Claims.(*jwtClaim)
			if !ok || !token.Valid {
				return response.Error(ctx, models.ErrorInvalidToken)
			}

			if claims.ExpiresAt.Time.Before(time.Now()) {
				return response.Error(ctx, models.ErrorExpiredToken)
			}

			if strings.ToLower(claims.Status) != "active" || strings.ToLower(claims.Status) == "inactive" {
				return response.Error(ctx, models.ErrorLoggedOut)
			}

			ctx.Set("communityId", claims.CommunityId)
			ctx.Set("userType", claims.UserType)
			ctx.Set("roles", claims.Roles)
			return next(ctx)
		}
	}
}
