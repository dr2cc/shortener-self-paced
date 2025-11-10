// Package server is wrapper around built in http server.
package server

import (
	"app/internal/config"
	services "app/internal/usecase/shortener"
)

type Server interface {
	Run() error
	Shutdown() error
}

// func New(config *config.Config, ipChecker services.IPCheckerInterface, service *services.Shortener) (Server, error) {
func New(config *config.Config, service *services.Shortener) (Server, error) {
	// if config.EnableHTTPS {
	// 	return NewHTTPS(config, ipChecker, service)
	// } else {
	return NewHTTP(config, service)
	//}
}
