package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3UseSSL      bool
	FrontendURL   string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:      envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     envOr("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     envOr("JWT_SECRET", "dev-secret-change-me"),
		S3Endpoint:    envOr("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:   envOr("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:   envOr("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:      envOr("S3_BUCKET", "ravenchat"),
		S3UseSSL:      envOr("S3_USE_SSL", "false") == "true",
		FrontendURL:   envOr("FRONTEND_URL", "http://localhost:5173"),
	}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
