// Copyright 2025 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/weblog"
)

const (
	ExitUsage    = iota + 1 // ExitUsage indicates a usage error.
	ExitConfig              // ExitConfig indicates a config error.
	ExitLog                 // ExitLog indicates a log error.
	ExitHandler             // ExitHandler indicates a handler error.
	ExitServer              // ExitServer indicates a server error.
	ExitTemplate            // ExitTemplate indicates a template error.
)

func main() {
	// Check command line for config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := webapp.LoadConfigFromJSON(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get config:", err)
		os.Exit(ExitConfig)
	}

	// Validate config.
	missingFields, err := cfg.MissingFields()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to validate config:", err)
		os.Exit(ExitConfig)
	}
	if len(missingFields) != 0 {
		fmt.Fprintln(os.Stderr, "Missing fields in config:", missingFields)
		os.Exit(ExitConfig)
	}

	// Initialize logging.
	err = weblog.Init(cfg.Log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
	}

	// Show config in log.
	slog.Info("using config", slog.Any("config", cfg))

	// Create the web app.
	app, err := webapp.New(webapp.WithName(cfg.App.Name))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating new handler:", err)
		os.Exit(ExitHandler)
	}

	// Create new ServeMux for HTTP requests and add routes and middleware.
	mux := http.NewServeMux()
	RegisterRoutes(mux, app)
	handler := AddMiddleware(mux)

	// Create the web server.
	srv, err := cfg.Server.Create(handler)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitServer)
	}

	// Create a new context.
	ctx := context.Background()

	// Start the web server.
	err = srv.Run(ctx)
	if err != nil {
		slog.Error("error starting server", slog.Any("err", err))
		fmt.Fprintln(os.Stderr, "Error starting server:", err)
		os.Exit(ExitServer)
	}
}
