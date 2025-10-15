package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	MongoURI      string
	JWTSecret     string
	EncryptionKey string
	ServiceName   string
	ServicePort   string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-here"),
		ServiceName:   getEnv("SERVICE_NAME", "unknown-service"),
		ServicePort:   getEnv("SERVICE_PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
