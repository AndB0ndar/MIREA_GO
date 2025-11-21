package config

import "os"

type Config struct {
	Port       string
	SocketPath string
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

	return Config{
		Port:       ":" + port,
		SocketPath: socket,
	}
}
