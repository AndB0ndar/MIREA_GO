package config

import (
	"os"
	"time"
)

type Config struct {
	Port       string
	SocketPath string
	Database   DatabaseConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Load() Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	socket := os.Getenv("SOCKET_PATH")
	if socket == "" {
		socket = "/tmp/app.sock"
	}

	dbURL := os.Getenv("NOTES_POSTGRESQL_URL")
	if dbURL == "" {
		dbURL = "postgres://user:pass@localhost:5433/notes?sslmode=disable"
	}

	maxOpenConns := 20
	maxIdleConns := 10
	connMaxLifetime := 30 * time.Minute
	connMaxIdleTime := 5 * time.Minute

	return Config{
		Port:       ":" + port,
		SocketPath: socket,
		Database: DatabaseConfig{
			URL:             dbURL,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
	}
}
