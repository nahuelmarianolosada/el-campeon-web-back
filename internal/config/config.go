package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	ServerPort       int
	ServerEnv        string
	JWTSecretKey     string
	JWTRefreshSecret string
	JWTExpiryHours   int

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// MercadoPago
	MercadopagoAccessToken string
	MercadopagoPublicKey   string

	// API URLs
	APIBaseURL string
}

func Load() *Config {
	return &Config{
		ServerPort:             getEnvInt("PORT", 8080),
		ServerEnv:              getEnv("ENV", "development"),
		JWTSecretKey:           getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		JWTRefreshSecret:       getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
		JWTExpiryHours:         getEnvInt("JWT_EXPIRY_HOURS", 24),
		DBHost:                 getEnv("DB_HOST", "localhost"),
		DBPort:                 getEnvInt("DB_PORT", 3306),
		DBUser:                 getEnv("DB_USER", "root"),
		DBPassword:             getEnv("DB_PASSWORD", "password"),
		DBName:                 getEnv("DB_NAME", "el_campeon_web"),
		MercadopagoAccessToken: getEnv("MERCADOPAGO_ACCESS_TOKEN", ""),
		MercadopagoPublicKey:   getEnv("MERCADOPAGO_PUBLIC_KEY", ""),
		APIBaseURL:             getEnv("API_BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
