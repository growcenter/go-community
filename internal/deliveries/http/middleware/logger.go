package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type bodyDumpResponseWriter struct {
    io.Writer
    http.ResponseWriter
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func (m *Middleware) LoggingMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(ctx echo.Context) error {
            // Request logging
            req := ctx.Request()
            res := ctx.Response()

            // Read request body
            var reqBody []byte
            if req.Body != nil {
                reqBody, _ = io.ReadAll(req.Body)
                req.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Restore body for further processing
            }

            // Dump response body
            resBody := new(bytes.Buffer)
            mw := io.MultiWriter(res.Writer, resBody)
            writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: res.Writer}
            ctx.Response().Writer = writer

            start := time.Now()

            // Process request
            err := next(ctx)

            // Log request and response details
            stop := time.Now()
            latency := stop.Sub(start)

            logger.Info("request",
                zap.String("method", req.Method),
                zap.String("uri", req.RequestURI),
                zap.String("remote_ip", ctx.RealIP()),
                zap.String("host", req.Host),
                zap.String("user_agent", req.UserAgent()),
                zap.Any("headers", req.Header),
                zap.ByteString("request_body", reqBody),
                zap.Int("status", res.Status),
                zap.ByteString("response_body", resBody.Bytes()),
                zap.Duration("latency", latency),
                zap.Time("start_time", start),
                zap.Time("end_time", stop),
                zap.String("error", errToString(err)),
            )

            return err
        }
    }
}

func errToString(err error) string {
    if err != nil {
        return err.Error()
    }
    return ""
}
