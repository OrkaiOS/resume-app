package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/handlers"
)

// Recovery recovers from panics in the request path and returns a 500 with
// the standard error envelope rather than Gin's default text response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatusJSON(500, handlers.Failure("INTERNAL", "internal server error"))
			}
		}()
		c.Next()
	}
}
