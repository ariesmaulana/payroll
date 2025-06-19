package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	Debug      bool
	JWTSecret  string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	// is okay if .env file not found, we can read directly on os level

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "myapp"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		Debug:      getEnv("DEBUG", "false") == "true",
		JWTSecret:  getEnv("SECRET_KEY", "2387126871hsadhajksdh89789"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
