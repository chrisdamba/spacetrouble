package main

import (
	"context"
	"fmt"
	"github.com/chrisdamba/spacetrouble/internal/api"
	"github.com/chrisdamba/spacetrouble/internal/ports"
	"github.com/chrisdamba/spacetrouble/internal/repository"
	"github.com/chrisdamba/spacetrouble/internal/service"
	"github.com/chrisdamba/spacetrouble/internal/utils"
	"github.com/chrisdamba/spacetrouble/pkg/config"
	"github.com/chrisdamba/spacetrouble/pkg/health"
	"github.com/chrisdamba/spacetrouble/pkg/spacex"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"syscall"
)

type App struct {
	config *config.Config
	server *http.Server
	db     *pgxpool.Pool
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Initialize(ctx context.Context) error {
	if err := a.setupDatabase(ctx); err != nil {
		return fmt.Errorf("database setup failed: %w", err)
	}

	if err := a.setupServer(); err != nil {
		return fmt.Errorf("server setup failed: %w", err)
	}

	return nil
}

func (a *App) setupDatabase(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(a.config.Database.DSN())
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	a.db = pool
	return nil
}

func (a *App) setupServer() error {
	services := a.setupServices()
	router := a.setupRouter(services)

	a.server = &http.Server{
		Addr:         a.config.Server.Address,
		Handler:      router,
		WriteTimeout: a.config.Server.WriteTimeout,
		ReadTimeout:  a.config.Server.ReadTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	return nil
}

type Services struct {
	BookingService ports.BookingService
}

func (a *App) setupServices() Services {
	repo := repository.NewBookingRepository(a.db)
	spaceXClient := spacex.NewClient(
		spacex.WithBaseURL(a.config.SpaceX.BaseURL),
	)

	return Services{
		BookingService: service.NewBookingService(repo, spaceXClient),
	}
}

func (a *App) setupRouter(services Services) http.Handler {
	router := http.NewServeMux()
	const versionPrefix = "/v1"

	router.HandleFunc(versionPrefix+"/health", health.HealthGet())

	bookingHandler := utils.AllowedMethods(
		utils.AllowedContentTypes(
			api.BookingHandler(services.BookingService),
			"application/json",
		),
		"POST", "GET", "DELETE",
	)
	router.HandleFunc(versionPrefix+"/bookings", bookingHandler)

	return router
}

func (a *App) Run(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-shutdown:
		log.Println("Starting graceful shutdown")
		return a.Shutdown(ctx)
	case <-ctx.Done():
		return a.Shutdown(ctx)
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if a.db != nil {
		a.db.Close()
	}

	return nil
}

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	app := NewApp(cfg)
	if err := app.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
