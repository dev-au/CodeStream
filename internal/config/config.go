package config

import (
	"os"
	"strconv"
	"time"

	"4d63.com/tz"
	"github.com/joho/godotenv"
)

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Port: mustInt("APPLICATION_HTTP_PORT"),
		},

		Application: AppConfig{
			Mode:                 os.Getenv("APPLICATION_MODE"),
			TimeZone:             getTimeZone(os.Getenv("TIMEZONE")),
			LogLevel:             os.Getenv("LOG_LEVEL"),
		},

		PostgresSQL: PostgreSQLConfig{
			User:     os.Getenv("POSTGRESQL_USER"),
			Password: os.Getenv("POSTGRESQL_PASSWORD"),
			Host:     os.Getenv("POSTGRESQL_HOST"),
			Port:     os.Getenv("POSTGRESQL_PORT"),
			Database: os.Getenv("POSTGRESQL_DATABASE"),
		},
	}

	return &cfg
}

func mustInt(key string) int {
	val := os.Getenv(key)
	n, err := strconv.Atoi(val)
	if err != nil {
		panic("Invalid integer for " + key)
	}
	return n
}

func getTimeZone(location string) *time.Location {
	timezone, err := tz.LoadLocation(location)
	if err != nil {
		panic("Invalid timezone for " + location)
	}
	return timezone
}
