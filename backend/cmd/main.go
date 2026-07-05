package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/config"
	"github.com/marco/resume-app/internal/handlers"
	"github.com/marco/resume-app/internal/middleware"
)

// Run loads configuration, wires the Gin router with middleware and routes,
// and starts the HTTP server. It is the single entrypoint for resume-app.
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
	)

	health := handlers.NewHealthHandler()
	metricsHandler := handlers.NewMetricsHandler()

	router.GET("/health", health.Health)
	router.GET("/metrics", metricsHandler.Metrics)

	log.Printf("resume-app server listening on :%s", cfg.Port)
	return router.Run(":" + cfg.Port)
}

func main() {
	if err := Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
