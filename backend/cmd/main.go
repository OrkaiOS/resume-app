package main

import (
	"embed"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/config"
	"github.com/marco/resume-app/internal/handlers"
	"github.com/marco/resume-app/internal/middleware"
)

func serveSPA(router *gin.Engine, f embed.FS) error {
	sub, err := fs.Sub(f, "static")
	if err != nil {
		return err
	}

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		data, err := fs.ReadFile(sub, path)
		if err == nil {
			ext := filepath.Ext(path)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			c.Data(http.StatusOK, contentType, data)
			return
		}

		data, err = fs.ReadFile(sub, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	return nil
}

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	metrics := middleware.NewMetrics()

	router := gin.New()
	router.Use(
		middleware.Recovery(),
		middleware.Logging(),
		metrics.Handler(),
		middleware.CORS(cfg.CORSAllowedOrigins),
	)

	health := handlers.NewHealthHandler()
	metricsHandler := handlers.NewMetricsHandler()

	router.GET("/health", health.Health)
	router.GET("/metrics", metricsHandler.Metrics)

	v1 := router.Group("/v1/api")

	orkaiHealth := handlers.NewOrkaiHealthHandler(cfg.OrkaiHealthURL)
	v1.GET("/health/orkai", orkaiHealth.CheckHealth)

	if hasFrontendFS {
		if err := serveSPA(router, prodFrontendFS); err != nil {
			return err
		}
	}

	log.Printf("resume-app server listening on :%s", cfg.Port)
	return router.Run(":" + cfg.Port)
}

func main() {
	if err := Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
