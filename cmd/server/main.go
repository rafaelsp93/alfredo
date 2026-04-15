package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rafaelsoares/alfredo/internal/app"
	"github.com/rafaelsoares/alfredo/internal/config"
	"github.com/rafaelsoares/alfredo/internal/database"
	"github.com/rafaelsoares/alfredo/internal/gcalendar"
	pethttp "github.com/rafaelsoares/alfredo/internal/petcare/adapters/primary/http"
	petmw "github.com/rafaelsoares/alfredo/internal/petcare/adapters/primary/http/middleware"
	petcaresqlite "github.com/rafaelsoares/alfredo/internal/petcare/adapters/secondary/sqlite"
	petsvc "github.com/rafaelsoares/alfredo/internal/petcare/service"
)

var version = "dev"

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Guard: refuse to start without a sufficiently strong API key
	if cfg.Auth.APIKey == "" {
		log.Fatal("APP_AUTH_API_KEY must be set — server refuses to start without authentication")
	}
	if len(cfg.Auth.APIKey) < 32 {
		log.Fatal("APP_AUTH_API_KEY must be at least 32 characters — use a cryptographically random value")
	}

	// 3. Init Zap logger with configurable level
	lvl, err := zapcore.ParseLevel(cfg.Log.Level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(lvl)
	zapLogger, err := zapCfg.Build()
	if err != nil {
		log.Fatalf("logger error: %v", err)
	}
	defer zapLogger.Sync() //nolint:errcheck

	// 4. Open SQLite (runs all migrations)
	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		zapLogger.Fatal("sqlite open failed", zap.Error(err))
	}
	defer db.Close() //nolint:errcheck

	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		zapLogger.Fatal("timezone load failed", zap.String("timezone", cfg.App.Timezone), zap.Error(err))
	}

	// 5. Calendar adapter (no-op when credentials are absent)
	var calendarAdapter app.CalendarPort
	if cfg.GCalendar.ClientID != "" && cfg.GCalendar.ClientSecret != "" && cfg.GCalendar.RefreshToken != "" {
		calendarAdapter, err = gcalendar.NewAdapter(context.Background(), gcalendar.AdapterConfig{
			ClientID:     cfg.GCalendar.ClientID,
			ClientSecret: cfg.GCalendar.ClientSecret,
			RefreshToken: cfg.GCalendar.RefreshToken,
		})
		if err != nil {
			zapLogger.Fatal("gcalendar init failed", zap.Error(err))
		}
	} else {
		calendarAdapter = gcalendar.NewNoopAdapter(zapLogger)
	}

	// 6. Pet-care repositories
	petRepo := petcaresqlite.NewPetRepository(db)
	vaccineRepo := petcaresqlite.NewVaccineRepository(db)
	treatmentRepo := petcaresqlite.NewTreatmentRepository(db)
	doseRepo := petcaresqlite.NewDoseRepository(db)
	txRunner := petcaresqlite.NewTxRunner(db)
	dbChecker := database.NewChecker(db)

	// 7. Pet-care services (pure CRUD — no side-effects)
	petService := petsvc.NewPetService(petRepo)
	vaccineService := petsvc.NewVaccineService(vaccineRepo, petRepo)
	treatmentService := petsvc.NewTreatmentService(treatmentRepo)
	doseService := petsvc.NewDoseService(doseRepo)

	// 8. Use Cases (orchestrate domain + calendar integration)
	petUC := app.NewPetUseCase(petService, txRunner, calendarAdapter, zapLogger)
	vaccineUC := app.NewVaccineUseCase(vaccineService, petService, txRunner, calendarAdapter, loc.String(), zapLogger)
	treatmentUC := app.NewTreatmentUseCase(treatmentService, doseService, petService, txRunner, calendarAdapter, loc.String(), zapLogger)

	// 9. Health aggregator
	healthAgg := app.NewHealthAggregator(map[string]app.HealthPinger{
		"sqlite": dbChecker,
	})
	healthHandler := pethttp.NewHealthHTTPHandler(healthAgg)

	// 10. HTTP handlers
	petHandler := pethttp.NewPetHandler(petUC)
	vaccineHandler := pethttp.NewVaccineHandler(vaccineUC, loc)
	treatmentHandler := pethttp.NewTreatmentHandler(treatmentUC, loc)

	// 11. Echo instance with global middleware
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(petmw.RequestLogger(zapLogger))
	e.Use(echomw.Recover())
	e.Use(echomw.BodyLimit("1M"))

	// Custom error handler: return consistent JSON and avoid leaking internal details.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		msg := "internal_error"
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			msg = http.StatusText(he.Code)
		}
		if !c.Response().Committed {
			_ = c.JSON(code, map[string]string{"error": msg})
		}
	}

	// 12. Public routes — no auth (health checks from Traefik/Docker)
	public := e.Group("/api/v1")
	public.GET("/health", healthHandler.Health)

	// 13. Protected routes — API key required
	protected := e.Group("/api/v1")
	protected.Use(petmw.APIKeyAuth(cfg.Auth.APIKey))
	petHandler.Register(protected)
	vaccineHandler.Register(protected)
	treatmentHandler.Register(protected)

	// 14. Start server with graceful shutdown
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	zapLogger.Info("alfredo starting", zap.String("addr", addr), zap.String("version", version))

	srv := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Error("shutdown error", zap.Error(err))
	}
}
