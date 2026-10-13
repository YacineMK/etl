package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr         string
	DatabaseURL       string
	WeatherURI        string
	RequestsPerSecond int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		RedisAddr:         getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nasa_space?sslmode=disable"),
		WeatherURI:        getEnv("WEATHER_URI", "https://archive-api.open-meteo.com/v1/archive"),
		RequestsPerSecond: getEnvInt("REQUESTS_PER_SECOND", 8),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
