// Package config handles environment-based configuration for the application.
package config

import "os"

// StorageType represents the backing storage engine.
type StorageType string

const (
	// StorageMemory uses in-memory stores, suitable for local development.
	StorageMemory StorageType = "memory"

	// StorageRedis uses Redis-backed stores, suitable for production.
	StorageRedis StorageType = "redis"
)

// Config holds all application configuration values.
type Config struct {
	// Port is the HTTP server listening port.
	Port string
	// Storage is the resolved storage backend to use.
	Storage StorageType
	// RedisURL is the Redis server address (only used when Storage is StorageRedis).
	RedisURL string
}

// Load reads environment variables and resolves the configuration.
// The storage type is determined by TWIX_ENV: "production" maps to Redis,
// everything else defaults to in-memory.
func Load() Config {
	env := getEnv("TWIX_ENV", "development")

	storage := StorageMemory
	if env == "production" {
		storage = StorageRedis
	}

	return Config{
		Port:     getEnv("PORT", "8080"),
		Storage:  storage,
		RedisURL: getEnv("REDIS_URL", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
