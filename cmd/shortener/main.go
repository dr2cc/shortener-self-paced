package main

import (
	"app/internal/app"
	"app/internal/config"
	"fmt"
	"log"
	"os"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// TODO: обработка ошибки при старте

	//app.Run(cfg)

	// Run
	if err := app.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
