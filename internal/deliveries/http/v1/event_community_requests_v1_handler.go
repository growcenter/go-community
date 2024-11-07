package v1

import (
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/deliveries/http/middleware"
	"go-community/internal/models"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type EventCommunityRequestHandler struct {
	usecase usecases.EventCommunityRequestUsecase
}

func NewEventCommunityRequestHandler(api *echo.Group, u usecases.EventCommunityRequestUsecase, c *config.Configuration) {
	handler := &EventCommunityRequestHandler{usecase: u}

	// Define event community request routes
	endpoint := api.Group("/community-request")
	endpoint.Use(middleware.UserMiddleware(c))
	endpoint.POST("/", handler.CreateRequest)
	endpoint.GET("/:id", handler.GetRequestByID)
	endpoint.GET("/account/:account_number", handler.GetRequestsByAccountNumber)
}

// CreateRequest handles creating a new event community request
func (h *EventCommunityRequestHandler) CreateRequest(ctx echo.Context) error {
	var request models.CreateEventCommunityRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	// Validate the request data
	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	// Use case to create the request
	newRequest, err := h.usecase.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	// Return the created request
	return response.Success(ctx, http.StatusCreated, newRequest.ToResponse())
}

// GetRequestByID handles fetching a request by its ID
func (h *EventCommunityRequestHandler) GetRequestByID(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	// Use case to get the request by ID
	request, err := h.usecase.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return response.Error(ctx, err)
	}

	// Return the request details
	return response.Success(ctx, http.StatusOK, request.ToResponse())
}

// GetRequestsByAccountNumber handles fetching all requests by account number
func (h *EventCommunityRequestHandler) GetRequestsByAccountNumber(ctx echo.Context) error {
	accountNumber := ctx.Param("account_number")

	// Use case to get all requests by account number
	requests, err := h.usecase.GetAllByAccountNumber(ctx.Request().Context(), accountNumber)
	if err != nil {
		return response.Error(ctx, err)
	}

	// Prepare the response
	var responseRequests []models.EventCommunityRequestResponse
	for _, req := range requests {
		responseRequests = append(responseRequests, *req.ToResponse())
	}

	// Return the list of requests
	return response.SuccessList(ctx, http.StatusOK, len(responseRequests), responseRequests)
}
