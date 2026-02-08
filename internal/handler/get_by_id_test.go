package handler

import (
	"app/internal/config"
	"app/internal/service"
	"net/http"
	"testing"
)

func TestHandler_Redirect(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		service *service.Service
		config  *config.Config
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			h.Redirect(tt.w, tt.r)
		})
	}
}
