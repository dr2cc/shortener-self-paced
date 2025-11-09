// Package config contains configuration for application.
package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	DSN string `yaml:"dsn" env-required:"true"`
	// готовлюсь к миграциям
	MigrationsPath string
	HTTPServer     `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	// User        string        `yaml:"user" env-required:"true"`
	// Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func NewConfig() (*Config, error) {

	configPath := os.Getenv("CONFIG_PATH") //"./config/config.yml" // Если путь задать жестко, не работает отладка
	if configPath == "" {
		log.Fatal("env CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg, nil
}
