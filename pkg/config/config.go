package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	UploadDir  string
	MediaURL   string
	ServerPort string
	ServerURL  string
	LogDir     string
	LogLevel   string
	JWTSecret  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading it. Relying on environment variables.")
	}

	return &Config{
		DBHost:     mustGetEnv("DB_HOST"),
		DBUser:     mustGetEnv("DB_USER"),
		DBPassword: mustGetEnv("DB_PASSWORD"),
		DBName:     mustGetEnv("DB_NAME"),
		DBPort:     mustGetEnv("DB_PORT"),
		ServerPort: mustGetEnv("SERVER_PORT"),
		ServerURL:  mustGetEnv("SERVER_URL"),
		LogDir:     mustGetEnv("LOG_DIR"),
		LogLevel:   mustGetEnv("LOG_LEVEL"),
		MediaURL:   mustGetEnv("MEDIA_URL"),
		UploadDir:  mustGetEnv("UPLOAD_DIR"),
		JWTSecret:  mustGetEnv("JWT_SECRET"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}
