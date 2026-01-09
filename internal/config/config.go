// Package config contains configuration for application.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	ServerAddress string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	FilePath      string `json:"file_storage_path"`
	DatabaseDSN   string `json:"database_dsn"`
	ConfigPath    string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Env:           "local", // Окружение - local, dev или prod,в первую очередь для логгера
		ServerAddress: "",
		BaseURL:       "",
		FilePath:      "",
		DatabaseDSN:   "",
		ConfigPath:    "",
	}

	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.BaseURL, "b", "", "resulting shortened URL")

	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn for connecting to postgres")
	flag.StringVar(&cfg.ConfigPath, "c", "", "config path")

	flag.Parse()

	configFromFile, err := cfg.parseConfigFile(cfg.ConfigPath)
	if err != nil {
		return &Config{}, err
	}

	// Считываем конфигурацию в такой последовательности:
	// - из флагов командной строки, - переменных окружения, - файла конфигурации, - значение по умолчанию
	cfg.ServerAddress = priorityLine(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), configFromFile.ServerAddress, ":8080")
	cfg.BaseURL = priorityLine(cfg.BaseURL, os.Getenv("BASE_URL"), configFromFile.BaseURL, "http://localhost:8080")
	cfg.FilePath = priorityLine(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), configFromFile.FilePath)
	cfg.DatabaseDSN = priorityLine(cfg.DatabaseDSN, os.Getenv("DATABASE_DSN"), configFromFile.DatabaseDSN)

	return cfg, nil
}

func priorityLine(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}

// func priorityBool(bools ...bool) bool {
// 	for _, boolVar := range bools {
// 		if boolVar {
// 			return true
// 		}
// 	}
// 	return false
// }

func (c *Config) parseConfigFile(configPath string) (Config, error) {
	if configPath == "" {
		return Config{}, nil
	}

	f, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found at: %s", configPath)
		}
		return Config{}, err
	}

	configFromFile := Config{}

	err = json.Unmarshal(f, &configFromFile)
	return configFromFile, err
}
