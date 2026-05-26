package config

import "time"

type Config struct {
	HTTP        HTTPConfig
	Application AppConfig
	PostgresSQL PostgreSQLConfig
}

type HTTPConfig struct {
	Port int
}

type AppConfig struct {
	Mode     string
	TimeZone *time.Location
	LogLevel string
}

type PostgreSQLConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}
