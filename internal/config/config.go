// Package config contains configuration for application.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// const KeySize = 2 * aes.BlockSize //nolint:gomnd

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	ServerAddress string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	FilePath      string `json:"file_storage_path"`
	DatabaseDSN   string `json:"database_dsn"`
	ConfigPath    string
	// MigrationsPath string
	// TrustedSubnet  string `json:"trusted_subnet"`
	// EncryptionKey  []byte
	// EnableHTTPS    bool `json:"enable_https"`
}

// New reads the configuration from the command line flags,
// environment variables and a configuration file (with priority).
func NewConfig() (*Config, error) {
	// key := []byte(os.Getenv("ENCRYPTION_KEY"))
	// if len(key) == 0 {
	// 	key = generateNewEncryptionKey()
	// }

	cfg := &Config{
		Env:           "local", // Окружение - local, dev или prod,в первую очередь для логгера
		ServerAddress: "",
		BaseURL:       "",
		FilePath:      "",
		DatabaseDSN:   "",
		ConfigPath:    "",
		// EncryptionKey:  key,
		// MigrationsPath: getEnv("MIGRATIONS_PATH", "file://internal/app/storage/migrations/"),
		// EnableHTTPS:    false,
		// TrustedSubnet:  "",
	}

	flag.StringVar(&cfg.ServerAddress, "a", "", "host to listen on")
	flag.StringVar(&cfg.BaseURL, "b", "", "base url")
	// "pip" или другой, это файл (работает и без расширения),
	// который будет (должен) находиться в каталоге ./cmd/shortener/
	// если его не указать- "", то "хранилище" в файле (iter9) не будет создаваться
	flag.StringVar(&cfg.FilePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn for connecting to postgres")
	flag.StringVar(&cfg.ConfigPath, "c", "", "config path")
	// flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet (CIDR notation)")
	// flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable https")

	flag.Parse()

	configFromFile, err := cfg.parseConfigFile(cfg.ConfigPath)
	if err != nil {
		return &Config{}, err
	}

	cfg.ServerAddress = coalesceStrings(cfg.ServerAddress, os.Getenv("SERVER_ADDRESS"), configFromFile.ServerAddress, ":8080")
	cfg.BaseURL = coalesceStrings(cfg.BaseURL, os.Getenv("BASE_URL"), configFromFile.BaseURL, "http://localhost:8080")
	cfg.FilePath = coalesceStrings(cfg.FilePath, os.Getenv("FILE_STORAGE_PATH"), configFromFile.FilePath)
	cfg.DatabaseDSN = coalesceStrings(cfg.DatabaseDSN, os.Getenv("DATABASE_DSN"), configFromFile.DatabaseDSN)
	// cfg.EnableHTTPS = coalesceBool(cfg.EnableHTTPS, os.Getenv("ENABLE_HTTPS") == "true", configFromFile.EnableHTTPS)
	// cfg.TrustedSubnet = coalesceStrings(cfg.TrustedSubnet, os.Getenv("TRUSTED_SUBNET"), configFromFile.TrustedSubnet, "127.0.0.1/24")

	return cfg, nil
}

// // generateNewEncryptionKey generates a random key of the specified KeySize.
// func generateNewEncryptionKey() []byte {
// 	randomGenerator := random.TrulyRandomGenerator{}
// 	randomKey, err := randomGenerator.GenerateRandomBytes(KeySize)
// 	if err != nil {
// 		randomKey = make([]byte, KeySize)
// 	}
// 	return randomKey
// }

func coalesceStrings(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}

// func coalesceBool(bools ...bool) bool {
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

// // If the environment variable exists, return it, otherwise return the fallback value.
// func getEnv(key, fallback string) string {
// 	if value, ok := os.LookupEnv(key); ok {
// 		return value
// 	}
// 	return fallback
// }
