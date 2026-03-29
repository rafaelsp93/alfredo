package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/rafaelsoares/alfredo/internal/logger"
)

// APIKeyAuth returns middleware that enforces API key authentication.
// Accepted headers (first match wins):
//   - Authorization: Bearer <key>
//   - X-Api-Key: <key>
//
// Returns 401 if the key is missing or does not match validKey.
func APIKeyAuth(validKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := extractAPIKey(c.Request())
			if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) != 1 {
				logger.FromEcho(c).Warn("auth: rejected",
					zap.String("client_ip", c.RealIP()),
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			return next(c)
		}
	}
}

func extractAPIKey(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.Header.Get("X-Api-Key")
}
