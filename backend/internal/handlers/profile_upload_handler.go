package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
)

const maxUploadSize = 10 << 20

type profileParser interface {
	ParsePDF(r io.Reader) (*models.Profile, error)
	ParseMarkdown(r io.Reader) (*models.Profile, error)
}

type ProfileUploadHandler struct {
	parser profileParser
}

func NewProfileUploadHandler(parser profileParser) *ProfileUploadHandler {
	return &ProfileUploadHandler{parser: parser}
}

func (h *ProfileUploadHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "file is required: "+err.Error())
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))

	var profile *models.Profile
	switch ext {
	case ".pdf":
		profile, err = h.parser.ParsePDF(file)
	case ".md", ".markdown":
		profile, err = h.parser.ParseMarkdown(file)
	default:
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "unsupported file type: "+ext+" (use .pdf or .md)")
		return
	}

	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(profileToResponse(*profile)))
}
