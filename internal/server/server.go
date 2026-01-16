package server

import (
	"net/http"
	"time"
)

func New(handler http.Handler, port string) *http.Server {

	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
}
