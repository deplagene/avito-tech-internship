package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
	Env         string
}

func InitConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/avito_trainee?sslmode=disable"),
		Env:         getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if val, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fallback
		}

		return i

	}

	return fallback
}
