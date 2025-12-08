package integration

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "app/internal/db"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestWithTestContainers(t *testing.T) {
    ctx := context.Background()

    // Запускаем контейнер с PostgreSQL
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("notes_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second)),
    )
    require.NoError(t, err)

    // Гарантируем остановку контейнера
    defer func() {
        if err := pgContainer.Terminate(ctx); err != nil {
            t.Errorf("failed to terminate container: %s", err)
        }
    }()

    // Получаем строку подключения
    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    // Подключаемся к БД
    database, err := sql.Open("pgx", connStr)
    require.NoError(t, err)
    defer database.Close()

    // Проверяем подключение
    err = database.Ping()
    require.NoError(t, err)

    // Инициализируем схему
    err = db.InitSchema(database)
    require.NoError(t, err)

    // Проверяем, что таблица создана
    var tableExists bool
    err = database.QueryRow(`SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = 'notes'
    )`).Scan(&tableExists)
    require.NoError(t, err)
    require.True(t, tableExists, "Table 'notes' should exist")

    // Проверяем, что можем вставить и получить данные
    _, err = database.Exec(
        "INSERT INTO notes (title, content) VALUES ($1, $2)",
        "TestContainers Note",
        "This note was created via testcontainers",
    )
    require.NoError(t, err)

    var count int
    err = database.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
    require.NoError(t, err)
    require.Equal(t, 1, count)

    t.Log("TestContainers test passed: database is ready and operations work")
}

