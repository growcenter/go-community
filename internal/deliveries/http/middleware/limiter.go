package middleware

import (
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

var (
	// Create a rate limiter with a limit of 5 requests per minute.
	limiter = rate.NewLimiter(5, 1) // 5 requests per minute with a burst of 1
	mtx     sync.Mutex
)

func (m *Middleware) RateLimiterMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if ctx.Request().Method != http.MethodPost {
				// If the method is not POST, continue to the next handler
				return next(ctx)
			}

			// Acquire a token from the rate limiter
			mtx.Lock()
			defer mtx.Unlock()

			if !limiter.Allow() {
				// Return a custom error response if rate limit is exceeded
				return response.Error(ctx, models.ErrorRateLimiterExceeds)
			}

			// Continue to the next handler if allowed
			return next(ctx)
		}
	}
}
