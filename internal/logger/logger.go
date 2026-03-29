package logger

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "log_logger"
const errorKey contextKey = "log_error"

// Set stores the request-scoped logger in the Echo context.
// Called by RequestLogger middleware.
func Set(c echo.Context, l *zap.Logger) {
	c.Set(string(loggerKey), l)
}

// FromEcho retrieves the request-scoped logger from the Echo context.
// Returns zap.NewNop() if none has been set (safe fallback for tests and direct handler calls).
func FromEcho(c echo.Context) *zap.Logger {
	if l, ok := c.Get(string(loggerKey)).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}

// SetError stores the domain error string in the Echo context so
// the RequestLogger middleware can include it in the completion log.
func SetError(c echo.Context, errStr string) {
	c.Set(string(errorKey), errStr)
}

// GetError retrieves the stored domain error string.
// Called by RequestLogger after the handler returns.
func GetError(c echo.Context) string {
	if s, ok := c.Get(string(errorKey)).(string); ok {
		return s
	}
	return ""
}
