package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"net/http"
	"strings"
)

type ConfigHandler struct {
	conf *config.Configuration
}

func NewConfigHandler(api *echo.Group, c *config.Configuration) {
	handler := &ConfigHandler{conf: c}

	departmentEndpoint := api.Group("/departments")
	departmentEndpoint.GET("", handler.GetDepartments)

	campusEndpoint := api.Group("/campuses")
	campusEndpoint.GET("", handler.GetCampuses)
}

func (ch *ConfigHandler) GetDepartments(ctx echo.Context) error {
	department := ch.conf.Department

	var departments []models.DepartmentsResponse
	for key, value := range department {
		departments = append(departments, models.DepartmentsResponse{
			Type:           models.TYPE_DEPARTMENT,
			DepartmentCode: strings.ToUpper(key),
			DepartmentName: value,
		})
	}

	return response.SuccessList(ctx, http.StatusOK, len(departments), departments)
}

func (ch *ConfigHandler) GetCampuses(ctx echo.Context) error {
	campus := ch.conf.Campus

	var campuses []models.CampusesResponse
	for key, value := range campus {
		campuses = append(campuses, models.CampusesResponse{
			Type:       models.TYPE_CAMPUS,
			CampusCode: strings.ToUpper(key),
			CampusName: value,
		})
	}

	return response.SuccessList(ctx, http.StatusOK, len(campuses), campuses)
}
