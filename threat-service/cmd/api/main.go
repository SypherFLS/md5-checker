package main

import (
	"fmt"
	"net/http"
	"os"

	"ksp/internal/bootstrap"
)

func main() {
	cfg, err := bootstrap.ConfigGet()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config initialization failed:", err)
		os.Exit(1)
	}
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "application initialization failed:", err)
		os.Exit(1)
	}

	defer func() {
		if err := app.Close(); err != nil {
			app.Logger.Error("failed to close application", "error", err)
		}
	}()

	if err := http.ListenAndServe(cfg.Server.Port, app.Router); err != nil {
		app.Logger.Error("HTTP server stopped", "error", err)
		os.Exit(1)
	}
}