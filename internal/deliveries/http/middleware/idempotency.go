package middleware

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

var (
	processedRequests = make(map[string]bool)
	mu                sync.Mutex
)

func (m *Middleware) IdempotencyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Method == http.MethodGet {
			return next(ctx)
		}

		requestID := ctx.Request().Header.Get("X-Request-ID")

		if requestID == "" {
			return response.Error(ctx, models.ErrorEmptyRequestID)
		}

		mu.Lock()
		if processedRequests[requestID] {
			mu.Unlock()
			return response.Error(ctx, models.ErrorProcessedRequestID)
		}
		processedRequests[requestID] = true
		mu.Unlock()

		return next(ctx)
	}
}
