package v1

//
//import (
//	"go-community/internal/config"
//	"go-community/internal/deliveries/http/common/response"
//	"go-community/internal/usecases"
//	"net/http"
//
//	"github.com/labstack/echo/v4"
//)
//
//type DeparmentHandler struct {
//	config *config.Configuration
//}
//
//func NewDepartmentHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
//	handler := &DeparmentHandler{}
//
//	// Define campus routes
//	endpoint := api.Group("/departments")
//	endpoint.GET("", handler.GetAllDepartment)
//}
//
//func (dh *DeparmentHandler) GetAllDepartment(ctx echo.Context) error {
//	department := dh.config.Department
//	return response.SuccessList(ctx, http.StatusOK, len(department), department)
//}
