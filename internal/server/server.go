// В пакете server создаем "оболочку" (wrapper)
// вокруг встроенного метода http.Server
package server

import (
	"context"
	"net/http"
	"time"
)

//// Решение от ypgo с последующим выбором типа сервера (http или https)
//// Раскритиковано ментором.
// type Server interface {
// 	Run() error
// 	Shutdown() error
// }

// // Вызывается из app
// func New(config *config.Config, router chi.Router) (Server, error) {
// 	// Здесь будем запускать HTTP и HTTPS (инкремент 21)
// // NewHTTP в ypgo по пути http.NewHTTP
// 	return NewHTTP(config, router)
// }

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
