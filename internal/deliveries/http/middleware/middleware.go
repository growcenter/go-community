package middleware

import (
	"github.com/labstack/echo/v4"
)

type Middleware struct {
	e *echo.Echo
}

func New(e *echo.Echo) Middleware {
	return Middleware{
		e: e,
	}
}
