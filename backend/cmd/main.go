package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/config"
	"github.com/marco/resume-app/internal/handlers"
	"github.com/marco/resume-app/internal/middleware"
	"github.com/marco/resume-app/internal/services"
	"github.com/marco/resume-app/internal/store"
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

	db, err := store.InitDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("cmd.Run: %w", err)
	}
	defer db.Close()

	profileStore := store.NewSQLiteProfileStore(db)
	opportunityStore := store.NewSQLiteOpportunityStore(db)
	resumeStore := store.NewSQLiteResumeStore(db)
	coverLetterStore := store.NewSQLiteCoverLetterStore(db)
	artifactStore := store.NewSQLiteArtifactStore(db)

	profileSvc := services.NewProfileService(profileStore)
	opportunitySvc := services.NewOpportunityService(opportunityStore)
	resumeSvc := services.NewResumeService(resumeStore)
	coverLetterSvc := services.NewCoverLetterService(coverLetterStore)
	artifactSvc := services.NewArtifactService(artifactStore)

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

	profileHandler := handlers.NewProfileHandler(profileSvc)
	v1.GET("/profile", profileHandler.Get)
	v1.PUT("/profile", profileHandler.Upsert)

	opportunityHandler := handlers.NewOpportunityHandler(opportunitySvc)
	v1.GET("/opportunities", opportunityHandler.List)
	v1.POST("/opportunities", opportunityHandler.Create)
	v1.GET("/opportunities/:id", opportunityHandler.Get)
	v1.PUT("/opportunities/:id", opportunityHandler.Update)
	v1.DELETE("/opportunities/:id", opportunityHandler.Delete)
	v1.PUT("/opportunities/:id/archive", opportunityHandler.SetArchived)

	resumeHandler := handlers.NewResumeHandler(resumeSvc)
	v1.GET("/opportunities/:id/resume", resumeHandler.GetByOpportunity)
	v1.PUT("/opportunities/:id/resume", resumeHandler.Upsert)

	coverLetterHandler := handlers.NewCoverLetterHandler(coverLetterSvc)
	v1.GET("/opportunities/:id/cover-letter", coverLetterHandler.GetByOpportunity)
	v1.PUT("/opportunities/:id/cover-letter", coverLetterHandler.Upsert)

	artifactHandler := handlers.NewArtifactHandler(artifactSvc)
	v1.GET("/tools/artifacts", artifactHandler.List)
	v1.POST("/tools/artifacts", artifactHandler.Create)
	v1.GET("/tools/artifacts/:id", artifactHandler.Get)
	v1.DELETE("/tools/artifacts/:id", artifactHandler.Delete)

	if hasFrontendFS {
		if err := serveSPA(router, prodFrontendFS); err != nil {
			return err
		}
	}

	log.Printf("resume-app server listening on :%s", cfg.Port)
	return router.Run(":" + cfg.Port)
}

func main() {
	serve, exit := resolveCommand(os.Args[1:])
	if !serve {
		if exit != 0 {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		}
		printUsage()
		if exit != 0 {
			os.Exit(1)
		}
		return
	}
	if err := Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func resolveCommand(args []string) (serve bool, exit int) {
	if len(args) == 0 || args[0] == "serve" {
		return true, 0
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return false, 0
	}
	return false, 1
}

func printUsage() {
	printUsageTo(os.Stdout)
}

func printUsageTo(w io.Writer) {
	fmt.Fprintln(w, `orkai-resume — a local-first resume builder

Usage:
  orkai-resume [command]

Commands:
  serve    Start the resume-app server (default)
  help     Print this help message`)
}
