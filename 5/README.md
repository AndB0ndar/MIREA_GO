# Практическое задание 5. Подключение к PostgreSQL через database/sql

Студент группы *ЭФМО-02-25 Бондарь Андрей Ренатович*

## Описание

**Цели:**

- Установить и настроить PostgreSQL локально;
- Подключиться к БД из Go с помощью database/sql и драйвера PostgreSQL;
- Выполнить параметризованные запросы `INSERT` и `SELECT`;
- Корректно работать с context, пулом соединений и обработкой ошибок.

## Инициализация проекта

```bash
mkdir -p app/cmd/server app/internal/db app/internal/repository
cd app
go mod init app
go get github.com/jackc/pgx/v5 v5.5.0
go get github.com/joho/godotenv v1.5.1
```

## Структура проекта

```
app
├── cmd
│   └── server
│       └── main.go
├── go.mod
├── go.sum
└── internal
    ├── db
    │   └── db.go
    └── repository
        └── repository.go
```

## Реализация

### internal/db/db.go

Подключение к PostgreSQL и настройка пула соединений:

- Использование драйвера pgx через стандартный интерфейс database/sql;
- Настройка параметров пула соединений (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`);
- Проверка соединения через `PingContext` с таймаутом.

```go
func OpenDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return nil, err
    }

    log.Println("Connected to PostgreSQL")
    return db, nil
}
```

### internal/repository/repository.go

CRUD-операции с задачами:

- `CreateTask` - создание задачи с возвратом ID;
- `ListTasks` - получение всех задач;
- `ListDone` - фильтрация по статусу выполнения;
- `FindByID` - поиск задачи по идентификатору;
- `CreateMany` - массовая вставка через транзакцию;
- `UpdateTaskStatus` - обновление статуса задачи.

```go
func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
    var id int
    const q = `INSERT INTO tasks (title) VALUES ($1) RETURNING id;`
    err := r.DB.QueryRowContext(ctx, q, title).Scan(&id)
    return id, err
}

func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
    tx, err := r.DB.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    const q = `INSERT INTO tasks (title) VALUES ($1);`
    for _, title := range titles {
        if _, err := tx.ExecContext(ctx, q, title); err != nil {
            return err
        }
    }
    return tx.Commit()
}
```

### cmd/server/main.go

Основная логика приложения:

- Загрузка конфигурации из переменных окружения;
- Подключение к базе данных;
- Демонстрация всех CRUD-операций;
- Работа с контекстами для таймаутов.

```go
func main() {
    _ = godotenv.Load()
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatalf("Not found variable DATABASE_URL!")
    }

    database, err := db.OpenDB(dsn)
    if err != nil {
        log.Fatalf("openDB error: %v", err)
    }
    defer database.Close()

    repo := repository.NewRepo(database)

    // Выполнение операций с задачами
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    // Создание отдельных задач
    titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
    for _, title := range titles {
        id, err := repo.CreateTask(ctx, title)
        if err != nil {
            log.Fatalf("CreateTask error: %v", err)
        }
        log.Printf("Inserted task id=%d (%s)", id, title)
    }
}
```

## Настройка базы данных

### Создание БД и таблицы

```sql
CREATE DATABASE todo;
\c todo

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tasks (title) VALUES ('Первая задача из psql');
SELECT * FROM tasks;
```

![sql](img/sql.png)

### Конфигурация подключения (.env)

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/todo?sslmode=disable
```

## Тестирование

### Запуск приложения

```bash
go run .
```

![run](img/run.png)

### Проверка операций

Приложение демонстрирует:

1. Создание отдельных задач - постинг трех задач в базу данных.
2. Массовая вставка - добавление нескольких задач в рамках одной транзакции.
3. Обновление статуса - изменение статуса задачи на "выполнено".
4. Получение всех задач - вывод полного списка задач.
5. Фильтрация по статусу - получение только выполненных задач.
6. Поиск по ID - получение конкретной задачи по идентификатору.

## Особенности реализации

- Параметризованные запросы - использование плейсхолдеров `$1`, `$2` для защиты от SQL-инъекций;
- Пул соединений - оптимизация производительности через настройку пула;
- Контексты - использование `context.WithTimeout` для контроля времени выполнения запросов;
- Транзакции - обеспечение атомарности при массовых операциях;
- Обработка ошибок - корректная обработка и логирование ошибок на всех этапах.

