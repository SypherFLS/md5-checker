package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: app.Router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("HTTP server stopped", "error", err)
			_ = app.Close()
			os.Exit(1)
		}
	case sig := <-stop:
		app.Logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			app.Logger.Error("graceful shutdown failed", "error", err)
		}
	}

	if err := app.Close(); err != nil {
		app.Logger.Error("failed to close application", "error", err)
		os.Exit(1)
	}
}
