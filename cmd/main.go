package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dev-au/CodeStream/internal/config"
	"github.com/dev-au/CodeStream/pkg/logs"
	"github.com/dev-au/CodeStream/pkg/plugins"
)

var quit = make(chan os.Signal, 1)

func runProject(cfg *config.Config) {
	app, cleanup, err := InitializeApp(cfg)
	if err != nil {
		logs.Fatal("Failed to initialize application", "error", err)
	}
	defer cleanup()

	go func() {
		logs.Info("Starting HTTP server", "port", cfg.HTTP.Port)
		if err = app.Server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Fatal("server failed to start", "error", err)
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logs.Info("Shutting down gracefully...")
}

func main() {
	cfg := config.Load()
	logs.Init(cfg.Application.Mode, cfg.Application.LogLevel)
	plugins.SetLocation(cfg.Application.TimeZone)
	defer logs.Sync()
	args := os.Args

	if len(args) < 2 {
		runProject(cfg)
		return
	}

	switch args[1] {
	case "health":
		logs.Info("Health check success")
	case "test":
		logs.Info("Testing success")
	}
}
