package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBCharset  string

	JWTSecret       string
	JWTExpiresHours int
}

func Load() *Config {
	_ = godotenv.Load()

	expiry, err := strconv.Atoi(getEnv("JWT_EXPIRES_HOURS", "8"))
	if err != nil || expiry < 1 {
		expiry = 8
	}

	return &Config{
		AppName:         getEnv("APP_NAME", "AACS Plantation API"),
		AppPort:         getEnv("APP_PORT", "8081"),
		AppEnv:          getEnv("APP_ENV", "development"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          getEnv("DB_NAME", "aacs_plantation"),
		DBCharset:       getEnv("DB_CHARSET", "utf8mb4"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTExpiresHours: expiry,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
