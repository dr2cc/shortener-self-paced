// Package server is wrapper around built in http server.
package server

import (
	"app/internal/config"
	handlers "app/internal/controller/http"
	services "app/internal/usecase/shortener"
	"context"
	"net/http"
	"time"
)

type HTTP struct {
	server *http.Server
}

// Run starts http server.
func (s *HTTP) Run() error {
	return s.server.ListenAndServe()
}

func (s *HTTP) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

// func NewHTTP(config *config.Config, ipChecker services.IPCheckerInterface, service *services.Shortener) (Server, error) {
func NewHTTP(config *config.Config, service *services.Shortener) (Server, error) {
	httpServer := &http.Server{
		Addr:              config.ServerAddress,
		Handler:           handlers.NewRouter(service, config),
		ReadHeaderTimeout: 1 * time.Second,
	}
	server := &HTTP{
		server: httpServer,
	}
	return server, nil
}
