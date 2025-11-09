// Package httpserver implements HTTP server.
package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 5 * time.Second
)

// New -.
func New(addr string, router *chi.Mux, log *slog.Logger) *http.Server {
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// Логика web-сервера
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	return httpServer
}
