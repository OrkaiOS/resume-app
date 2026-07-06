package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/store"
)

func respondError(c *gin.Context, status int, code string, msg string) {
	c.JSON(status, Failure(code, msg))
}

func mapError(err error) (int, string) {
	if errors.Is(err, store.ErrNotFound) {
		return http.StatusNotFound, ErrCodeNotFound
	}
	if errors.Is(err, store.ErrConflict) {
		return http.StatusConflict, ErrCodeConflict
	}
	return http.StatusInternalServerError, ErrCodeInternal
}
