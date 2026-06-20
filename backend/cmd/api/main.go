// Package main is the entry point for the Ventiqra API service.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/auth"
	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
	"github.com/YASSERRMD/Ventiqra/backend/internal/db"
	"github.com/YASSERRMD/Ventiqra/backend/internal/logger"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
	"github.com/YASSERRMD/Ventiqra/backend/internal/server"
)

// shutdownTimeout is the maximum time allowed to drain in-flight requests.
const shutdownTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ventiqra-api: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	// Connect to PostgreSQL and apply migrations.
	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelConnect()

	pool, err := db.Connect(connectCtx, cfg.DSN(), db.DefaultPoolConfig())
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close(pool)
	log.Info("database connected")

	migrateCtx, cancelMigrate := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelMigrate()
	if n, err := db.Migrate(migrateCtx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	} else if n > 0 {
		log.Info("migrations applied", "count", n)
	}

	// Stop on interrupt or termination signals.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	opts := []server.Option{server.WithDB(pool)}

	if tm, err := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL); err != nil {
		log.Warn("auth disabled: JWT secret not configured", "error", err)
	} else {
		baseRepo := repository.New(pool)
		opts = append(opts,
			server.WithAuth(repository.NewUserRepo(baseRepo), tm),
			server.WithCompany(repository.NewCompanyRepo(baseRepo)),
			server.WithSim(repository.NewSimStateRepo(baseRepo)),
			server.WithProducts(repository.NewProductRepo(baseRepo)),
			server.WithEmployees(repository.NewEmployeeRepo(baseRepo)),
			server.WithLaunches(repository.NewLaunchRepo(baseRepo)),
			server.WithCustomers(repository.NewCustomerRepo(baseRepo)),
		)
		log.Info("auth, company, simulation, product, employee, launch, and customer services enabled")
	}

	srv := server.New(cfg, log, opts...)

	// Serve until the server stops on its own.
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	}

	// Gracefully drain in-flight requests.
	stop() // restore default signal behavior for a second Ctrl-C
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("ventiqra-api stopped")
	return nil
}
