package config

import (
	"os"
	"strconv"
)

type Config struct {
	DB_DSN     string
	BcryptCost int
	Addr       string
	SocketPath string
}

func Load() Config {
	cost := 12
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			cost = parsed
		}
	}

	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	socketPath := os.Getenv("SOCKET_PATH")
	if socketPath == "" {
		socketPath = "/tmp/app.sock"
	}

	return Config{
		DB_DSN:     os.Getenv("POSTGRESQL_DSN"),
		BcryptCost: cost,
		Addr:       addr,
		SocketPath: socketPath,
	}
}
