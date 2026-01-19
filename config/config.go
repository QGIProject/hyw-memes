package config

import (
	"os"
)

type Config struct {
	AdminPassword string
	JWTSecret     string
	DatabasePath  string
	UploadDir     string
	Port          string
}

var AppConfig *Config

func Load() {
	AppConfig = &Config{
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		DatabasePath:  getEnv("DATABASE_PATH", "./memes.db"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		Port:          getEnv("PORT", "3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
