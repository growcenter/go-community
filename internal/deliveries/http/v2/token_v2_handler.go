package v2

import (
	"go-community/internal/common"
	"go-community/internal/config"
	"go-community/internal/constants"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type TokenHandler struct {
	conf    *config.Configuration
	auth    *authorization.Auth
	usecase *usecases.Usecases
}

func NewTokenHandler(api *echo.Group, a *authorization.Auth, c *config.Configuration, u *usecases.Usecases) {
	handler := &TokenHandler{conf: c, auth: a, usecase: u}

	endpoint := api.Group("/tokens")
	endpoint.Use(middleware.RefreshMiddleware(c, u))
	endpoint.GET("", handler.Refresh)
	endpoint.POST("", handler.RefreshPost)
}

// Refresh godoc
// @Summary Generate Tokens
// @Description Generate both Access and Refresh Token
// @Tags tokens
// @Accept json
// @Produce json
// @Param Cookie path string true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.TokensResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/tokens [get]
func (th *TokenHandler) Refresh(ctx echo.Context) error {
	id := ctx.Get("id").(string)
	if common.IsValidUUID(id) {
		tokens, err := th.auth.GenerateTokens(id, constants.TYPE_GUEST, constants.ROLE_GUEST)
		if err != nil {
			response.Error(ctx, err)
		}

		return response.SuccessList(ctx, http.StatusCreated, 2, tokens.ToGenerateTokens())
	} else {
		user, err := th.usecase.User.GetRBAC(ctx.Request().Context(), id)
		if err != nil {
			return response.Error(ctx, err)
		}

		if user == nil {
			return response.Error(ctx, models.ErrorUserNotFound)
		}

		tokens, err := th.auth.GenerateTokens(user.CommunityId, user.UserTypes, user.Roles)
		if err != nil {
			response.Error(ctx, err)
		}

		return response.SuccessList(ctx, http.StatusCreated, 2, tokens.ToGenerateTokens())
	}
}

// RefreshPost godoc
// @Summary Generate Tokens using POST
// @Description Generate both Access and Refresh Token using POST method
// @Tags tokens
// @Accept json
// @Produce json
// @Param Cookie path string true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.List{data=[]models.TokensResponse} "Response indicates that the request succeeded and the new resources has been created and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v2/tokens [post]
func (th *TokenHandler) RefreshPost(ctx echo.Context) error {
	header := ctx.Request().Header.Get("X-Refresh-Token")
	if header == "" {
		tokens, err := th.auth.GenerateTokens(ctx.Get("id").(string), constants.TYPE_GUEST, constants.ROLE_GUEST)
		if err != nil {
			response.Error(ctx, err)
		}

		return response.SuccessList(ctx, http.StatusCreated, 2, tokens.ToGenerateTokens())
	} else {
		user, err := th.usecase.User.GetRBAC(ctx.Request().Context(), ctx.Get("id").(string))
		if err != nil {
			return response.Error(ctx, err)
		}

		if user == nil {
			return response.Error(ctx, models.ErrorUserNotFound)
		}

		tokens, err := th.auth.GenerateTokens(user.CommunityId, user.UserTypes, user.Roles)
		if err != nil {
			response.Error(ctx, err)
		}

		return response.SuccessList(ctx, http.StatusCreated, 2, tokens.ToGenerateTokens())
	}
}
