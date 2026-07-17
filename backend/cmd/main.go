package main

import (
	"context"
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
	"github.com/marco/resume-app/internal/llm"
	"github.com/marco/resume-app/internal/middleware"
	"github.com/marco/resume-app/internal/orkai"
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

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision Browser auto-open is gated on hasFrontendFS (prod build only); make dev uses Vite which owns the frontend, so no auto-open there. Headless env (SSH/DOCKER/no DISPLAY) skips silently.
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
	onboardingStore := store.NewSQLiteOnboardingStore(db)

	profileSvc := services.NewProfileService(profileStore)
	opportunitySvc := services.NewOpportunityService(opportunityStore)
	resumeSvc := services.NewResumeService(resumeStore)
	coverLetterSvc := services.NewCoverLetterService(coverLetterStore)
	artifactSvc := services.NewArtifactService(artifactStore)
	onboardingSvc := services.NewOnboardingServiceWithProfiles(onboardingStore, profileStore)

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

	profileParser := services.NewProfileParser()
	profileLLMParser := services.NewProfileLLMParser()
	profileUploadHandler := handlers.NewProfileUploadHandler(profileParser, profileLLMParser, profileSvc, onboardingSvc)
	v1.POST("/profile/upload", profileUploadHandler.Upload)

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

	orkaiClient := orkai.NewOrkaiClient(cfg.OrkaiMCPURL, cfg.OrkaiMCPToken)
	shellSvc := services.NewShellService()
	orkaiSearchSvc := services.NewOrkaiSearchService(orkaiClient, onboardingStore)
	sessionSvc := services.NewSessionService(orkaiClient, onboardingStore)
	toolsSvc := services.NewToolsService(shellSvc, orkaiSearchSvc, profileSvc)
	pdfSvc := services.NewPDFService(cfg.OutputDir)
	toolRegistry := services.NewToolRegistry(shellSvc, orkaiClient, onboardingStore, profileSvc, artifactSvc, sessionSvc, pdfSvc, opportunitySvc, resumeSvc, coverLetterSvc)
	toolsHandler := handlers.NewToolsHandler(toolsSvc)
	v1.POST("/tools/shell", toolsHandler.Shell)
	v1.POST("/tools/orkai-search", toolsHandler.OrkaiSearch)
	v1.GET("/tools/profile", toolsHandler.Profile)

	onboardingHandler := handlers.NewOnboardingHandler(onboardingSvc)
	v1.POST("/onboarding/llm-config", onboardingHandler.SaveLLMConfig)
	v1.POST("/onboarding/profile", onboardingHandler.SaveProfile)
	v1.GET("/onboarding/status", onboardingHandler.GetStatus)

	orkaiSetupSvc := services.NewOrkaiSetupService(orkaiClient, onboardingStore)
	orkaiSetupHandler := handlers.NewOrkaiSetupHandler(orkaiSetupSvc, profileSvc)
	v1.POST("/onboarding/orkai-setup", orkaiSetupHandler.StartSetup)
	v1.GET("/onboarding/orkai-setup/status", orkaiSetupHandler.GetStatus)

	systemPromptSvc := services.NewSystemPromptService(onboardingStore, profileStore, opportunityStore, orkaiClient, sessionSvc)

	llmClient := createLLMClient(cfg, onboardingStore)
	chatAgent := services.NewChatAgentService(llmClient, systemPromptSvc, toolRegistry)
	chatAgent.SetSessionSaver(sessionSvc)
	chatHandler := handlers.NewChatHandler(chatAgent)
	v1.POST("/chat", chatHandler.Stream)

	if hasFrontendFS {
		if err := serveSPA(router, prodFrontendFS); err != nil {
			return err
		}
		appURL := "http://localhost:" + cfg.Port
		go waitAndOpenBrowser(cfg.Port, appURL)
	}

	log.Printf("resume-app server listening on :%s", cfg.Port)
	return router.Run(":" + cfg.Port)
}

func createLLMClient(cfg config.Config, onboardingStore store.OnboardingStore) llm.Client {
	state, err := onboardingStore.Get(context.Background())
	provider := cfg.LLMProvider
	model := cfg.LLMModel
	apiKey := cfg.LLMAPIKey
	if err == nil && state.LLMProvider != "" {
		provider = state.LLMProvider
		model = state.LLMModel
		apiKey = state.LLMAPIKey
	}
	return llm.NewClient(provider, model, apiKey)
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
