package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets CORS headers based on the
// comma-separated allowedOrigins string. It never sets * as the origin;
// if the request's Origin header is not in the allowed list the header is
// omitted. Preflight OPTIONS requests are answered with 204 and the same
// headers.
func CORS(allowedOrigins string) gin.HandlerFunc {
	allowed := parseOrigins(allowedOrigins)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		for _, o := range allowed {
			if o == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
