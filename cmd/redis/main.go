package main

import (
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/burenotti/redis_impl/internal/config"
	"github.com/burenotti/redis_impl/internal/handler"
	"github.com/burenotti/redis_impl/internal/server"
	"github.com/burenotti/redis_impl/internal/service"
	"github.com/burenotti/redis_impl/internal/storage/memory"
)

var configPath string

func main() {
	notify := make(chan os.Signal, 1)
	defer close(notify)

	signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}))

	parseFlags()

	cfg := config.MustLoad(configPath)

	srv := initServer(logger, cfg)

	srvDone := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			srvDone <- err
			close(srvDone)
			return
		}
		logger.Info("Server stopped")
	}()

	select {
	case err := <-srvDone:
		if err != nil {
			logger.Error("Unexpected error while running server. Exiting.", "error", err)
			return
		}
	case <-notify:
	}

	if err := srv.Stop(cfg.Server.ShutdownTimeout); err != nil {
		if errors.Is(err, server.ErrStoppedAbnormally) {
			logger.Info("Server was stopped abnormally. Some connections were hung up")
		} else {
			logger.Error("Unexpected error while stopping server. Exiting.", "error", err)
		}
	} else {
		logger.Info("Server gracefully stopped")
	}
}

func initServer(logger *slog.Logger, cfg *config.Config) *server.Server {
	store := memory.New()
	handle := handler.New(func() *service.Controller {
		return service.New(store)
	})
	srv := server.Default(handle)
	srv.Host = cfg.Server.Host
	srv.Port = cfg.Server.Port
	srv.MaxConnections = cfg.Server.MaxConnections
	srv.Logger = logger
	return srv
}

func parseFlags() {
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()
	if !flag.Parsed() {
		panic("flags parsing failed")
	}
}
