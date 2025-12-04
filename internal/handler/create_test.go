package handlers_test

import (
	"app/internal/config"
	handlers "app/internal/handler"
	services "app/internal/usecase/shortener"
	"net/http"
	"testing"
)

func TestHandler_ShortenText(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		service *services.Shortener
		config  *config.Config
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewHandler(tt.service, tt.config)
			h.ShortenText(tt.w, tt.r)
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		service *services.Shortener
		config  *config.Config
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewHandler(tt.service, tt.config)
			h.ShortenAPI(tt.w, tt.r)
		})
	}
}
