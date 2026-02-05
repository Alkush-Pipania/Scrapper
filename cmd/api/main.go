package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alkush-Pipania/Scrapper/config"
	"github.com/Alkush-Pipania/Scrapper/internal/app"
	"github.com/Alkush-Pipania/Scrapper/internal/server"
	"github.com/Alkush-Pipania/Scrapper/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logger.Init(cfg)
	log.Info().Msg("logger initialized")

	container, err := app.NewContainer(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize dependencies")
	}

	// start consumer runs in seperate go routine (for link )
	app.StartConsumer(ctx, container)
	router := app.NewRouter(container)

	srv := server.New(router, cfg.Port, log)
	srv.Start()

	<-ctx.Done() // wait for the signal
	log.Info().Msg("shutdown signal received")

	// 1. Stop HTTP server (stop accepting requests)
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := container.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("dependecies shutdown failed")
	}

	// Shutdown done
	log.Info().Msg("graceful shutdown complete")

}
