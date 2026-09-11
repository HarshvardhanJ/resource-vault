package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	vault "github.com/HarshvardhanJ/resource-vault"
	"github.com/HarshvardhanJ/resource-vault/internal/catalog"
	"github.com/HarshvardhanJ/resource-vault/internal/config"
	"github.com/HarshvardhanJ/resource-vault/internal/db"
	"github.com/HarshvardhanJ/resource-vault/internal/reports"
	"github.com/HarshvardhanJ/resource-vault/internal/resources"
	"github.com/HarshvardhanJ/resource-vault/internal/storage"
	"github.com/HarshvardhanJ/resource-vault/internal/web"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "Run database migrations and exit")
	seedOnly := flag.Bool("seed", false, "Run database seed script and exit")
	flag.Parse()

	cfg := config.Load()
	var logHandler slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	if cfg.AppEnv == "production" { logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}) }
	logger := slog.New(logHandler); slog.SetDefault(logger)
	logger.Info("starting NITC Resource Vault", slog.String("env", cfg.AppEnv), slog.String("addr", cfg.AppAddr))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()

	// Local storage is retained for development/fixtures. Production configuration
	// should select the Internet Archive adapter once it is fully implemented.
	store, err := storage.NewLocalStorage(cfg.StorageDir, cfg.AppBaseURL)
	if err != nil { logger.Error("failed to initialize object storage", slog.Any("error", err)); os.Exit(1) }

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil { logger.Error("failed to connect to database", slog.Any("error", err)); os.Exit(1) }
	defer database.Close()

	if *migrateOnly || *seedOnly {
		if err := database.Migrate(context.Background()); err != nil { logger.Error("migration failed", slog.Any("error", err)); os.Exit(1) }
		if *seedOnly {
			if err := database.Seed(context.Background(), store); err != nil { logger.Error("seeding failed", slog.Any("error", err)); os.Exit(1) }
		}
		return
	}
	if err := database.Migrate(context.Background()); err != nil { logger.Error("startup migration failed", slog.Any("error", err)); os.Exit(1) }

	templatesFS, err := fs.Sub(vault.TemplatesFS, "templates")
	if err != nil { logger.Error("failed to locate embedded templates", slog.Any("error", err)); os.Exit(1) }
	renderer, err := web.NewRenderer(templatesFS)
	if err != nil { logger.Error("failed to parse templates", slog.Any("error", err)); os.Exit(1) }

	catalogRepo := catalog.NewRepository(database.Pool)
	catalogService := catalog.NewService(catalogRepo)
	resourcesRepo := resources.NewRepository(database.Pool)
	resourcesService := resources.NewService(resourcesRepo)
	catalogHandler := catalog.NewHandler(catalogService, resourcesService, renderer)
	academicUnitHandler := catalog.NewAcademicUnitHandler(catalog.NewAcademicUnitRepository(database.Pool), renderer)
	resourcesHandler := resources.NewHandler(resourcesService, store, renderer)
	reportsHandler := reports.NewHandler(database.Pool, renderer)

	r := chi.NewRouter()
	r.Use(web.RequestIDMiddleware)
	r.Use(web.LoggingMiddleware(logger))
	r.Use(web.RecoveryMiddleware(logger))
	r.Use(web.SecurityHeadersMiddleware)
	r.Use(middleware.Compress(5))

	staticFS, err := fs.Sub(vault.StaticFS, "static")
	if err != nil { logger.Error("failed to locate embedded static assets", slog.Any("error", err)); os.Exit(1) }
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		dbErr := database.Ping(r.Context())
		status, dbStatus, code := "ok", "connected", http.StatusOK
		if dbErr != nil { status, dbStatus, code = "degraded", "unreachable", http.StatusServiceUnavailable }
		w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": status, "db": dbStatus, "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})

	r.Get("/", catalogHandler.HandleHome)
	r.Get("/branches", catalogHandler.HandleHome)
	r.Get("/branches/{branch}", catalogHandler.HandleBranch)
	r.Get("/academic-units", academicUnitHandler.HandleList)
	r.Get("/courses/new", catalogHandler.HandleNewCourse)
	r.Post("/courses/new", catalogHandler.HandleCreateCourse)
	r.Get("/courses/{course}", catalogHandler.HandleCourse)
	r.Get("/courses/{course}/edit", catalogHandler.HandleEditCourse)
	r.Post("/courses/{course}/edit", catalogHandler.HandleUpdateCourse)
	r.Get("/resources/{id}", resourcesHandler.HandleResourceDetail)
	r.Get("/resources/{id}/download", resourcesHandler.HandleDownload)
	r.Get("/resources/{id}/preview", resourcesHandler.HandlePreview)
	r.Get("/search", resourcesHandler.HandleSearch)
	r.Get("/contribute", reportsHandler.HandleContributeInfo)
	r.Get("/reports/new", reportsHandler.HandleNewReport)
	r.Post("/reports", reportsHandler.HandleSubmitReport)

	srv := &http.Server{Addr: cfg.AppAddr, Handler: r, ReadTimeout: 15*time.Second, WriteTimeout: 30*time.Second, IdleTimeout: 60*time.Second}
	serverErrors := make(chan error, 1)
	go func() { logger.Info("HTTP server listening", slog.String("addr", cfg.AppAddr)); serverErrors <- srv.ListenAndServe() }()
	shutdown := make(chan os.Signal, 1); signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed { logger.Error("server error", slog.Any("error", err)); os.Exit(1) }
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second); defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil { logger.Error("failed to gracefully shutdown server", slog.Any("error", err)); _ = srv.Close() }
	}
}
