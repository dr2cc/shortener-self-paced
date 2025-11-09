package main

import (
	"app/internal/config"
	"app/internal/server"
	"log"
)

func main() {
	// 1️⃣ Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app := server.NewApp()
	app.Run(cfg)
}
