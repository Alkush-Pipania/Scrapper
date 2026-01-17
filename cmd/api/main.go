package api

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/Alkush-Pipania/Scrapper/config"
	"github.com/Alkush-Pipania/Scrapper/internal/app"
	"github.com/Alkush-Pipania/Scrapper/internal/server"
	"github.com/Alkush-Pipania/Scrapper/pkg/logger"
)

func main() {
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

	srv := server.New(router, cfg.Port)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to Start Server")
	}
}
