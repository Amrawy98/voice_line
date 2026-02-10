package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	GroqAPIKey    string
	MaxFileSizeMB int
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getEnvOrDefault("PORT", "8080"),
		GroqAPIKey:    os.Getenv("GROQ_API_KEY"),
		MaxFileSizeMB: getEnvIntOrDefault("MAX_FILE_SIZE_MB", 19),
	}

	if cfg.GroqAPIKey == "" {
		return Config{}, fmt.Errorf("GROQ_API_KEY is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
