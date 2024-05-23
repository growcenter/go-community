package health

import (
	"go-community/internal/models"
	"go-community/internal/pkg/response"
	"go-community/internal/usecases"
	"net/http"

	"github.com/labstack/echo/v4"
)

type healthHandler struct {
	usecase usecases.Usecases
}

func NewHealhHandler(api *echo.Group, uc usecases.Usecases) {
	hh := healthHandler{
		usecase: uc,
	}

	health := api.Group("/health")
	health.GET("", hh.Check)
}

func (hh healthHandler) Check(ctx echo.Context) error {
	if err := hh.usecase.Health.Check(ctx.Request().Context()); err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, models.Health{
		Type:   "health",
		Status: "Service up and running",
	})
}
