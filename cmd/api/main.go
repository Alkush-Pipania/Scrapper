package api

import (
	"context"
	"log"

	"github.com/Alkush-Pipania/Scrapper/config"
	"github.com/Alkush-Pipania/Scrapper/internal/app"
	"github.com/Alkush-Pipania/Scrapper/internal/server"
)

func main() {
	cfg := config.LoadEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container := app.NewContainer(ctx)

	router := app.NewRouter(container)

	srv := server.New(router, cfg.Port)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Server failed to start", err)
	}
}
