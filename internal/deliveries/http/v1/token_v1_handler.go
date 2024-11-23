package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"net/http"
)

type TokenHandler struct {
	conf *config.Configuration
	auth *authorization.Auth
}

func NewTokenHandler(api *echo.Group, a *authorization.Auth, c *config.Configuration) {
	handler := &TokenHandler{conf: c, auth: a}

	endpoint := api.Group("/tokens")
	endpoint.POST("", handler.Refresh)
}

func (th *TokenHandler) Refresh(ctx echo.Context) error {
	refresh, err := ctx.Cookie("refresh_token")
	if err != nil {
		response.Error(ctx, err)
	}

	claims, err := th.auth.ValidateRefresh(refresh.Value)
	if err != nil {
		response.Error(ctx, err)
	}

	newAccess, err := th.auth.GenerateAccessToken(claims.CommunityId, claims.UserType, claims.Roles, "active")
	if err != nil {
		response.Error(ctx, err)
	}

	res := models.GenerateAccessTokenResponse{Type: models.TYPE_ACCESS_TOKEN, AccessToken: newAccess}
	return response.Success(ctx, http.StatusCreated, res.ToCreateAccessToken())

}
