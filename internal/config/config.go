package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv         string
	AppAddr        string
	AppBaseURL     string
	DatabaseURL    string
	SessionSecret  string
	StorageDir     string
	MaxUploadMB    int
	LogLevel       string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GoogleHostedDomain string
	IAAccessKey        string
	IASecretKey        string
	IACollection       string
	IAIdentifierPrefix string
}

func Load() *Config {
	maxUploadMB, err := strconv.Atoi(getEnv("MAX_UPLOAD_MB", "25"))
	if err != nil || maxUploadMB <= 0 { maxUploadMB = 25 }
	return &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		AppAddr: getEnv("APP_ADDR", ":8080"),
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://archive:archive@postgres:5432/archive?sslmode=disable"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-change-in-production-min-32-chars"),
		StorageDir: getEnv("STORAGE_DIR", "./data/storage"),
		MaxUploadMB: maxUploadMB,
		LogLevel: getEnv("LOG_LEVEL", "info"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL: getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		GoogleHostedDomain: getEnv("GOOGLE_HOSTED_DOMAIN", "nitc.ac.in"),
		IAAccessKey: getEnv("IA_ACCESS_KEY", ""),
		IASecretKey: getEnv("IA_SECRET_KEY", ""),
		IACollection: getEnv("IA_COLLECTION", ""),
		IAIdentifierPrefix: getEnv("IA_IDENTIFIER_PREFIX", "nitc-resource-vault"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" { return val }
	return defaultVal
}
