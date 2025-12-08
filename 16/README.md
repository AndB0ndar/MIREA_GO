# Практическое задание 16. Интеграционное тестирование API. Использование Docker для тестовой БД

Студент группы *ЭФМО-02-25 Бондарь Андрей Ренатович*

## Описание

**Цели:**

- Освоить интеграционное тестирование REST API с проверкой всего стека приложения
- Научиться поднимать изолированную тестовую среду БД в Docker с использованием docker-compose
- Освоить инициализацию схемы БД, сидирование тестовых данных и очистку окружения
- Внедрить интеграционные проверки всех CRUD-эндпоинтов с валидацией статусов, заголовков, JSON-ответов и эффектов в БД

---

## Структура проекта

```
app
├── cmd
│   ├── api
│   │   └── main.go
│   └── server
│       └── main.go
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── integration
│   ├── notes_integration_test.go
│   └── notes_testcontainers_test.go
├── internal             
│   ├── core          
│   │   └── note.go        
│   ├── db
│   │   └── migrate.go
│   ├── http
│   │   ├── handlers
│   │   │   ├── notes.go
│   │   │   └── notes_test.go
│   │   ├── middleware
│   │   │   ├── cors.go
│   │   │   └── logger.go
│   │   └── router.go
│   ├── mathx
│   │   ├── mathx.go
│   │   └── mathx_test.go
│   ├── platform
│   │   ├── config
│   │   │   └── config.go
│   │   └── database
│   │       └── postgres.go
│   ├── repo
│   │   ├── interface.go
│   │   └── note_pg.go
│   ├── service
│   │   ├── repo.go
│   │   ├── service.go
│   │   └── service_test.go
│   └── stringsx
│       ├── stringsx.go
│       └── stringsx_test.go
└── Makefile
```

---

## Реализация

### Файл: `docker-compose.yml`

Конфигурация изолированной тестовой базы данных PostgreSQL.

**Ключевые компоненты:**

```yaml
version: '3.8'
services:
  postgres-test:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: notes_test
    ports:
      - "54321:5432"
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test -d notes_test"]
      interval: 5s
      timeout: 5s
      retries: 5
```

**Архитектурное значение:**

- Создание полностью изолированной тестовой среды
- Автоматическая инициализация схемы через SQL-скрипт
- Health check для гарантии готовности БД перед запуском тестов
- Отдельный порт (54321) для избежания конфликтов с локальной БД

### Файл: `init.sql`

Инициализация схемы БД с таблицами, индексами и триггерами.

**Ключевые компоненты:**

```sql
CREATE TABLE IF NOT EXISTS notes (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notes_title_gin 
ON notes USING GIN (to_tsvector('simple', title));

CREATE TRIGGER trg_notes_updated BEFORE UPDATE ON notes
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

**Архитектурное значение:**

- Декларативное описание схемы БД
- Создание индексов для оптимизации запросов
- Автоматическое обновление поля updated_at через триггеры
- Гарантированная согласованность данных

### Файл: `internal/db/migrate.go`

Программная инициализация и управление схемой БД.

**Ключевые компоненты:**

1. **Инициализация схемы:**
   
   ```go
   func InitSchema(db *sql.DB) error {
       queries := []string{
           `CREATE TABLE IF NOT EXISTS notes (...)`,
           `CREATE INDEX IF NOT EXISTS idx_notes_title_gin ...`,
           `CREATE OR REPLACE FUNCTION update_updated_at() ...`,
           `CREATE TRIGGER trg_notes_updated ...`,
       }
   
       for i, query := range queries {
           if _, err := db.Exec(query); err != nil {
               return fmt.Errorf("failed to execute query %d: %w", i, err)
           }
       }
       return nil
   }
   ```

2. **Очистка данных между тестами:**
   
   ```go
   func TruncateTables(db *sql.DB) error {
       if _, err := db.Exec("TRUNCATE TABLE notes RESTART IDENTITY CASCADE"); err != nil {
           return fmt.Errorf("failed to truncate tables: %w", err)
       }
       return nil
   }
   ```

3. **Сидирование тестовых данных:**
   
   ```go
   func SeedTestData(db *sql.DB) error {
       testNotes := []struct {
           title   string
           content string
       }{
           {"Test Note 1", "Content of test note 1"},
           {"Test Note 2", "Content of test note 2"},
       }
   
       for _, note := range testNotes {
           _, err := db.Exec(
               "INSERT INTO notes (title, content) VALUES ($1, $2)",
               note.title, note.content,
           )
           if err != nil {
               return err
           }
       }
       return nil
   }
   ```

**Архитектурное значение:**

- Централизованное управление миграциями
- Гарантированная идемпотентность операций (IF NOT EXISTS)
- Изоляция тестов через полную очистку данных
- Воспроизводимость тестовой среды

### Файл: `integration/notes_integration_test.go`

Основные интеграционные тесты API с реальной БД.

**Ключевые компоненты:**

1. **Настройка тестового окружения:**
   
   ```go
   func TestMain(m *testing.M) {
       dsn := os.Getenv("TEST_DATABASE_URL")
       if dsn == "" {
           dsn = "postgres://test:test@localhost:54321/notes_test?sslmode=disable"
       }
   
       testDB, err = sql.Open("pgx", dsn)
       if err != nil {
           panic(fmt.Sprintf("Failed to connect to test database: %v", err))
       }
   
       if err := db.InitSchema(testDB); err != nil {
           panic(fmt.Sprintf("Failed to initialize schema: %v", err))
       }
   
       code := m.Run()
       db.TruncateTables(testDB)
       os.Exit(code)
   }
   ```

2. **Тестирование создания заметки:**
   
   ```go
   func TestCreateNoteIntegration(t *testing.T) {
       setupTest(t)
   
       requestBody := `{"title": "Integration Test Note", "content": "Test content"}`
       req, err := http.NewRequest("POST", testServer.URL+"/notes", 
           bytes.NewBufferString(requestBody))
       require.NoError(t, err)
   
       resp, err := http.DefaultClient.Do(req)
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusCreated, resp.StatusCode)
   
       var responseNote core.Note
       body, _ := io.ReadAll(resp.Body)
       json.Unmarshal(body, &responseNote)
   
       assert.NotZero(t, responseNote.ID)
       assert.Equal(t, "Integration Test Note", responseNote.Title)
   }
   ```

3. **Тестирование получения заметки:**
   
   ```go
   func TestGetNoteIntegration(t *testing.T) {
       setupTest(t)
   
       var noteID int64
       err := testDB.QueryRow(
           "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
           "Test Note", "Test content",
       ).Scan(&noteID)
       require.NoError(t, err)
   
       resp, err := http.Get(fmt.Sprintf("%s/notes/%d", testServer.URL, noteID))
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusOK, resp.StatusCode)
   }
   ```

4. **Тестирование ошибки 404:**
   
   ```go
   func TestGetNoteNotFoundIntegration(t *testing.T) {
       setupTest(t)
   
       resp, err := http.Get(fmt.Sprintf("%s/notes/%d", testServer.URL, 99999))
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusNotFound, resp.StatusCode)
   }
   ```

5. **Тестирование обновления заметки:**
   
   ```go
   func TestUpdateNoteIntegration(t *testing.T) {
       setupTest(t)
   
       var noteID int64
       err := testDB.QueryRow(
           "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
           "Original Title", "Original Content",
       ).Scan(&noteID)
       require.NoError(t, err)
   
       updateBody := `{"title": "Updated Title", "content": "Updated Content"}`
       req, err := http.NewRequest("PATCH", 
           fmt.Sprintf("%s/notes/%d", testServer.URL, noteID),
           bytes.NewBufferString(updateBody))
       require.NoError(t, err)
   
       resp, err := http.DefaultClient.Do(req)
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusOK, resp.StatusCode)
   }
   ```

6. **Тестирование удаления заметки:**
   
   ```go
   func TestDeleteNoteIntegration(t *testing.T) {
       setupTest(t)
   
       var noteID int64
       err := testDB.QueryRow(
           "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
           "Note to Delete", "Content to delete",
       ).Scan(&noteID)
       require.NoError(t, err)
   
       req, err := http.NewRequest("DELETE", 
           fmt.Sprintf("%s/notes/%d", testServer.URL, noteID), nil)
       require.NoError(t, err)
   
       resp, err := http.DefaultClient.Do(req)
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusNoContent, resp.StatusCode)
   
       var count int
       err = testDB.QueryRow("SELECT COUNT(*) FROM notes WHERE id = $1", noteID).Scan(&count)
       require.NoError(t, err)
       assert.Equal(t, 0, count)
   }
   ```

7. **Тестирование получения всех заметок:**
   
   ```go
   func TestGetAllNotesIntegration(t *testing.T) {
       setupTest(t)
   
       notes := []struct {
           title   string
           content string
       }{
           {"Note 1", "Content 1"},
           {"Note 2", "Content 2"},
           {"Note 3", "Content 3"},
       }
   
       for _, note := range notes {
           _, err := testDB.Exec(
               "INSERT INTO notes (title, content) VALUES ($1, $2)",
               note.title, note.content,
           )
           require.NoError(t, err)
       }
   
       resp, err := http.Get(testServer.URL + "/notes")
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusOK, resp.StatusCode)
   
       var responseNotes []core.Note
       body, _ := io.ReadAll(resp.Body)
       json.Unmarshal(body, &responseNotes)
   
       assert.Len(t, responseNotes, 3)
   }
   ```

8. **Тестирование валидации:**
   
   ```go
   func TestCreateNoteValidationIntegration(t *testing.T) {
       setupTest(t)
   
       requestBody := `{"content": "Content without title"}`
       req, err := http.NewRequest("POST", testServer.URL+"/notes", 
           bytes.NewBufferString(requestBody))
       require.NoError(t, err)
   
       resp, err := http.DefaultClient.Do(req)
       require.NoError(t, err)
       defer resp.Body.Close()
   
       assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
   }
   ```

9. **Тестирование конкурентных запросов:**
   
   ```go
   func TestConcurrentRequestsIntegration(t *testing.T) {
       setupTest(t)
   
       const numRequests = 10
       errors := make(chan error, numRequests)
   
       for i := 0; i < numRequests; i++ {
           go func(id int) {
               requestBody := fmt.Sprintf(
                   `{"title": "Concurrent Note %d", "content": "Content %d"}`, id, id)
               req, _ := http.NewRequest("POST", testServer.URL+"/notes", 
                   bytes.NewBufferString(requestBody))
               resp, err := http.DefaultClient.Do(req)
               if err != nil {
                   errors <- err
                   return
               }
               defer resp.Body.Close()
   
               if resp.StatusCode != http.StatusCreated {
                   errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
                   return
               }
               errors <- nil
           }(i)
       }
   
       for i := 0; i < numRequests; i++ {
           err := <-errors
           assert.NoError(t, err)
       }
   }
   ```

**Архитектурное значение:**

- Полноценное тестирование всего стека приложения
- Проверка не только HTTP-ответов, но и состояния БД
- Изоляция тестов через очистку данных
- Тестирование граничных случаев и ошибок
- Проверка конкурентности

### Файл: `Makefile`

Упрощение запуска тестов и управления инфраструктурой.

**Ключевые компоненты:**

```makefile
up-test-db:
    docker-compose up -d postgres-test
    @sleep 5

down-test-db:
    docker-compose down -v

test-integration-docker: up-test-db
    TEST_DATABASE_URL="postgres://test:test@localhost:54321/notes_test?sslmode=disable" \
    go test -v -tags=integration ./integration/...
    $(MAKE) down-test-db
```

**Архитектурное значение:**

- Автоматизация рутинных операций
- Гарантированная последовательность действий
- Упрощение команд для разработчиков

---

## Тестирование

### Команды запуска

Способ 1: Использование Makefile

```bash
make test-integration-docker
```

Способ 2: Ручной запуск

```bash
docker-compose up -d postgres-test
sleep 5
TEST_DATABASE_URL="postgres://test:test@localhost:54321/notes_test?sslmode=disable" \
go test -v -tags=integration ./integration/...
docker-compose down -v
```

Способ 3: Запуск отдельных тестов

```bash
TEST_DATABASE_URL="postgres://test:test@localhost:54321/notes_test?sslmode=disable" \
go test -v -run TestCreateNoteIntegration ./integration/...
```

![test-integration-docker](./img/test-integration-docker.png)
![test-integration-docker-1](./img/test-integration-docker-1.png)
![test-integration-docker-2](./img/test-integration-docker-2.png)

### Проверка изоляции тестов

Проверка очистки данных между тестами

```bash
docker-compose exec postgres-test psql -U test -d notes_test -c "SELECT COUNT(*) FROM notes;"
```

![check-clear-db](./img/check-clear-db.png)

### Проверка состояния БД после тестов

Проверка структуры БД

```bash
docker-compose exec postgres-test psql -U test -d notes_test -c "\dt"
docker-compose exec postgres-test psql -U test -d notes_test -c "\di"
```

Проверка триггеров

```bash
docker-compose exec postgres-test psql -U test -d notes_test -c "
SELECT trigger_name, event_manipulation, action_statement 
    FROM information_schema.triggers 
    WHERE event_object_table = 'notes';
"
```

![check-state-db](./img/check-state-db.png)

---

## Заключение

В ходе интеграционного тестирования был полностью проверен
цикл работы REST API для заметок,
включая CRUD-операции, валидацию, обработку ошибок и конкурентность.
Тесты выполнялись в изолированной среде на базе Docker Compose,
где перед каждым запуском база данных полностью очищалась,
что обеспечило идемпотентность и надёжность проверок.

Основным результатом стало создание автоматизированной
и воспроизводимой системы тестирования всего стека приложения,
подтверждающей корректность взаимодействия API с реальной базой данных.
Были освоены ключевые навыки интеграционного тестирования и работы с Docker.
