package response

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
	"go-community/internal/common"
	"go-community/internal/models"
	"net/http"
	"time"
)

func Error(ctx echo.Context, err error) error {
	response := models.ErrorMapping(err)
	requestID, _ := ctx.Get("X-Request-Id").(string)
	if requestID == "" {
		requestID = uuid.New().String()
	}
	response.Metadata.RequestId = requestID

	timestamp, _ := ctx.Get("X-Timestamp").(string)
	if timestamp == "" {
		timestamp = common.Now().Format(time.RFC3339)
	}
	response.Metadata.Timestamp = timestamp

	return ctx.JSON(response.Code, response)
}

func ErrorValidation(ctx echo.Context, errors interface{}) error {
	requestID, _ := ctx.Get("X-Request-Id").(string)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	timestamp, _ := ctx.Get("X-Timestamp").(string)
	if timestamp == "" {
		timestamp = common.Now().Format(time.RFC3339)
	}

	response := models.Response{
		Code:    http.StatusUnprocessableEntity,
		Status:  "INVALID_INPUT",
		Message: "Validation failed for one or more fields.",
		Metadata: models.Metadata{
			RequestId: requestID,
			Timestamp: timestamp,
		},
	}
	if data, ok := errors.(*multierror.Error); ok {
		response.Errors = data.Errors
	}

	return ctx.JSON(response.Code, response)
}

func Success(ctx echo.Context, code int, response interface{}) error {
	return ctx.JSON(code, response)
}

func SuccessList(ctx echo.Context, code int, totalRows int, data interface{}) error {
	response := models.List{
		Type:      "collection",
		Data:      data,
		TotalRows: totalRows,
	}

	return ctx.JSON(code, response)
}

func SuccessListWithDetail(ctx echo.Context, code int, totalRows int, detail interface{}, data interface{}) error {
	response := models.ListWithDetail{
		Type:      "collection",
		Details:   detail,
		Data:      data,
		TotalRows: totalRows,
	}

	return ctx.JSON(code, response)
}

func SuccessPagination(ctx echo.Context, code int, pagination interface{}, data interface{}) error {
	response := models.Pagination{
		Type:           "collection",
		PaginationInfo: pagination,
		Data:           data,
	}

	return ctx.JSON(code, response)
}

func SuccessCursor(ctx echo.Context, code int, cursorInfo interface{}, data interface{}) error {
	response := models.Pagination{
		Type:           "collection",
		PaginationInfo: cursorInfo,
		Data:           data,
	}

	return ctx.JSON(code, response)
}

func SuccessDownload(ctx echo.Context, code int, contentType string, fileName string, data []byte) error {
	ctx.Response().Header().Set("Content-Type", contentType)
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename="+fileName)
	return ctx.Blob(http.StatusOK, contentType, data)
}

func SuccessV2(ctx echo.Context, code int, data interface{}) error {
	requestID, _ := ctx.Get("X-Request-Id").(string)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	timestamp, _ := ctx.Get("X-Timestamp").(string)
	if timestamp == "" {
		timestamp = common.Now().Format(time.RFC3339)
	}

	return ctx.JSON(http.StatusOK, models.Response{
		Code:    code,
		Status:  "RESPONSE_SUCCESS",
		Message: "Response has been successfully processed.",
		Data:    data,
		Metadata: models.Metadata{
			RequestId: requestID,
			Timestamp: timestamp,
		},
	})
}
