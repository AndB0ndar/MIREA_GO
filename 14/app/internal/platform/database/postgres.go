package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"app/internal/platform/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitPostgres(cfg config.DatabaseConfig) error {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns / 2)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database pool initialized: MaxConns=%d, MinConns=%d",
		cfg.MaxOpenConns, cfg.MaxIdleConns/2)

	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

func GetStats() string {
	if Pool == nil {
		return "Pool not initialized"
	}
	stats := Pool.Stat()
	return fmt.Sprintf("TotalConns: %d, AcquiredConns: %d, IdleConns: %d",
		stats.TotalConns(), stats.AcquiredConns(), stats.IdleConns())
}
