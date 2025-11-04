package db

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeps struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func ConnectMongo(ctx context.Context, url string, dbName string) (*MongoDeps, error) {
    clientOptions := options.Client().ApplyURI(url)

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
