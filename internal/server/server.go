package server

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Server struct {
	httpServer *http.Server
	logger     *zerolog.Logger
}

func New(handler http.Handler, port string, logger *zerolog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  1 * time.Minute,
		},
		logger: logger,
	}
}

func (s *Server) Start() {
	go func() {
		s.logger.Info().Msgf("HTTP server listening on %s", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal().Err(err).Msg("HTTP server crashed")
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("shutting down HTTP server")

	shutdownCtx, cancle := context.WithTimeout(ctx, 10*time.Second)
	defer cancle()

	return s.httpServer.Shutdown(shutdownCtx)
}
