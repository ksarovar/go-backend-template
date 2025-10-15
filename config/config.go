package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	MongoURI     string
	JWTSecret    string
	EncryptionKey string
}

// Load loads configuration from .env file and environment variables
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017/golang_backend"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "12345678901234567890123456789012"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
