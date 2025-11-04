package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeps struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongo(ctx context.Context, dsn string) (*MongoDeps, error) {
	clientOptions := options.Client().ApplyURI(dsn)

	// Устанавливаем таймаут для подключения
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Проверяем подключение
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	dbName, err := extractDatabaseNameFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract database name from DSN: %w", err)
	}

	database := client.Database(dbName)

	fmt.Println("Successfully connected to MongoDB!")
	return &MongoDeps{
		Client:   client,
		Database: database,
	}, nil
}

func (d *MongoDeps) Disconnect(ctx context.Context) error {
	return d.Client.Disconnect(ctx)
}

func extractDatabaseNameFromDSN(dsn string) (string, error) {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Путь в URL содержит слеш в начале, поэтому обрезаем его
	path := parsedURL.Path
	if path != "" && path[0] == '/' {
		path = path[1:]
	}

	// Если путь не пустой, используем его как имя базы данных
	if path != "" {
		return path, nil
	}

	// Если в пути нет имени БД, проверяем параметр authSource
	query := parsedURL.Query()
	if authSource := query.Get("authSource"); authSource != "" {
		return authSource, nil
	}

	// Если ничего не нашли, используем значение по умолчанию
	return "go", nil
}
