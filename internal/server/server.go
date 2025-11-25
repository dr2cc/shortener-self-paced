// В пакете server создаем "оболочку" (wrapper)
// вокруг встроенного метода http.Server
package server

import (
	"app/internal/config"

	"github.com/go-chi/chi/v5"
)

type Server interface {
	Run() error
	Shutdown() error
}

func New(config *config.Config, router chi.Router) (Server, error) {
	// Здесь будем запускать HTTP и HTTPS (инкремент 21)
	return NewHTTP(config, router)
}
