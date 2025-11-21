package config

import (
    "log"
    "os"
    "time"
)

type Config struct {
    Port            string
    SocketPath      string
    JWTSecret       []byte
    AccessTokenTTL  time.Duration
    RefreshTokenTTL time.Duration
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

    secret := os.Getenv("SECRET")
    if secret == "" {
        secret = "RNgu3fPJr1v2UyD+QEW2ZoXZtTNy+Ob1xdRhg05lQnQ="
        // log.Fatal("SECRET is required")
    }

    accessTTL := os.Getenv("ACCESS_TOKEN_TTL")
    if accessTTL == "" {
        accessTTL = "15m"
    }
    accessDur, err := time.ParseDuration(accessTTL)
    if err != nil {
        log.Fatal("bad ACCESS_TOKEN_TTL")
    }

    refreshTTL := os.Getenv("REFRESH_TOKEN_TTL")
    if refreshTTL == "" {
        refreshTTL = "168h" // 7 дней
    }
    refreshDur, err := time.ParseDuration(refreshTTL)
    if err != nil {
        log.Fatal("bad REFRESH_TOKEN_TTL")
    }

    return Config{
        Port:            ":" + port,
        SocketPath:      socket,
        JWTSecret:       []byte(secret),
        AccessTokenTTL:  accessDur,
        RefreshTokenTTL: refreshDur,
    }
}
