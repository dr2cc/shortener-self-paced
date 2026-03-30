package handler

import (
	"net/http"
	"testing"
)

func TestController_ShortenText(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *Controller
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ShortenText(tt.args.w, tt.args.r)
		})
	}
}

func TestController_ShortenAPI(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *Controller
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ShortenAPI(tt.args.w, tt.args.r)
		})
	}
}
