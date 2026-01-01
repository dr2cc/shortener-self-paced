// Package server is wrapper around built in http server.
package server

import (
	"app/internal/config"
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type HTTP struct {
	server *http.Server
}

// Run запускает http server.
func (s *HTTP) Run() error {
	return s.server.ListenAndServe()
}

func (s *HTTP) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

// Вызывается из server
func NewHTTP(config *config.Config, router chi.Router) (Server, error) {
	httpServer := &http.Server{
		Addr:              config.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: 1 * time.Second,
	}
	server := &HTTP{
		server: httpServer,
	}
	return server, nil
}
