// Package main is the entry point for the Ventiqra API service.
package main

import (
	"fmt"
	"os"

	"github.com/YASSERRMD/Ventiqra/backend/internal/config"
	"github.com/YASSERRMD/Ventiqra/backend/internal/logger"
	"github.com/YASSERRMD/Ventiqra/backend/internal/server"
)

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

	srv := server.New(cfg, log)
	return srv.ListenAndServe()
}
