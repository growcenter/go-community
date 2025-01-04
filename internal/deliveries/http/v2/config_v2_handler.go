package v2

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

// GetDepartments godoc
// @Summary Get Departments from Config
// @Description Get Departments from Config
// @Tags config
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.DepartmentsResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /api/v2/departments [get]
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

// GetCampuses godoc
// @Summary Get Campuses from Config
// @Description Get Campuses from Config
// @Tags config
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.List{data=[]models.CampusesResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=models.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /api/v2/campuses [get]
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
