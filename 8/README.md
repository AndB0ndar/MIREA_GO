# Практическое задание 8. Работа с MongoDB: подключение, создание коллекции, CRUD-операции

Студент группы *ЭФМО-02-25 Бондарь Андрей Ренатович*

## Описание

**Цели:**

- Понять базовые принципы документной БД MongoDB (документ, коллекция, BSON, \_id:ObjectID).
- Научиться подключаться к MongoDB из Go с использованием официального драйвера.
- Создать коллекцию, индексы и реализовать CRUD для одной сущности (например, notes).
- Отработать фильтрацию, пагинацию, обновления (в т.ч. частичные), удаление и обработку ошибок.

---

## Инициализация проекта

```bash
mkdir -p app/cmd/api app/internal/{db,notes}
cd app
go mod init app
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get github.com/go-chi/chi/v5
go get github.com/joho/godotenv
```

---

## Структура проекта

```
app
├── cmd
│   └── api
│       └── main.go
├── go.mod
├── go.sum
└── internal
    ├── db
    │   └── mongo.go
    └── notes
        ├── handler.go
        ├── model.go
        ├── repo.go
        └── repo_test.go
```

---

## Реализация

В этом разделе описывается реализация ключевых компонентов приложения
с объяснением назначения каждого файла и его роли в системе.

### Файл: `internal/db/mongo.go`

Инкапсуляция логики подключения к MongoDB и управление соединением.

**Ключевые компоненты:**

1. **Структура MongoDeps:**
   
   ```go
   type MongoDeps struct {
    Client   *mongo.Client
    Database *mongo.Database
   }
   ```
   
   Обертка вокруг MongoDB клиента и базы данных для dependency injection.

2. **Функция ConnectMongo:**
   
   ```go
   func ConnectMongo(ctx context.Context, uri, dbName string) (*MongoDeps, error) {
    opts := options.Client().ApplyURI(uri)
    cli, err := mongo.NewClient(opts)
    // Таймауты на подключение и проверку
    dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
   
    pingCtx, cancelPing := context.WithTimeout(ctx, 3*time.Second)
    defer cancelPing()
   }
   ```
   
   Инициализирует подключение к MongoDB с таймаутами для надежности.

**Архитектурное значение:**

- Централизует логику подключения к базе данных
- Обеспечивает обработку таймаутов и ошибок подключения
- Предоставляет единый интерфейс для работы с MongoDB

### Файл: `internal/notes/model.go`

Определение структуры данных и моделей для работы с заметками.

**Ключевые компоненты:**

1. **Структура Note:**
   
   ```go
   type Note struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title     string             `bson:"title" json:"title"`
    Content   string             `bson:"content" json:"content"`
    CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
   }
   ```
   
   Основная модель данных с тегами для BSON и JSON сериализации.

2. **Структуры запросов:**
   
   ```go
   type CreateNoteRequest struct {
    Title   string `json:"title" validate:"required,min=1,max=200"`
    Content string `json:"content" validate:"max=5000"`
   }
   ```
   
   DTO для валидации входящих запросов.

**Архитектурное значение:**

- Определяет доменную модель приложения
- Обеспечивает согласованность между слоями приложения
- Задает правила валидации данных

### Файл: `internal/notes/repo.go`

Реализация паттерна Repository для работы с данными в MongoDB.

**Ключевые компоненты:**

1. **Структура Repo:**
   
   ```go
   type Repo struct {
    col *mongo.Collection
   }
   ```
   
   Инкапсулирует работу с коллекцией MongoDB.

2. **Инициализация репозитория:**
   
   ```go
   func NewRepo(db *mongo.Database) (*Repo, error) {
    col := db.Collection("notes")
    // Создание уникального индекса на поле title
    _, err := col.Indexes().CreateOne(context.Background(), 
        mongo.IndexModel{
            Keys:    bson.D{{Key: "title", Value: 1}},
            Options: options.Index().SetUnique(true),
        })
   }
   ```
   
   Создает коллекцию и необходимые индексы при инициализации.

3. **CRUD операции:**
- **Create** - создание заметки с автоматической генерацией ID

- **ByID** - поиск по идентификатору с валидацией ObjectID

- **List** - получение списка с пагинацией и поиском

- **Update** - частичное обновление через оператор `$set`

- **Delete** - удаление с проверкой существования
4. **Обработка ошибок:**
   
   ```go
   var (
    ErrNotFound     = errors.New("note not found")
    ErrDuplicateKey = errors.New("duplicate title")
    ErrInvalidID    = errors.New("invalid id format")
   )
   ```
   
   Кастомные ошибки для единообразной обработки.

**Архитектурное значение:**

- Инкапсулирует всю логику работы с данными
- Предоставляет чистый интерфейс для бизнес-логики
- Обеспечивает повторное использование кода
- Облегчает тестирование через dependency injection

### Файл: `internal/notes/handler.go`

Реализация HTTP обработчиков и REST API для работы с заметками.

**Ключевые компоненты:**

1. **Структура Handler:**
   
   ```go
   type Handler struct{ repo *Repo }
   ```
   
   Обертка вокруг репозитория для HTTP взаимодействия.

2. **Маршрутизация:**
   
   ```go
   func (h *Handler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Get("/", h.list)
    r.Post("/", h.create)
    r.Get("/{id}", h.get)
    r.Patch("/{id}", h.patch)
    r.Delete("/{id}", h.del)
    return r
   }
   ```
   
   Настройка RESTful маршрутов с использованием роутера chi.

3. **Обработчики эндпоинтов:**
- **create** - создание заметки с валидацией

- **get** - получение заметки по ID

- **list** - список с пагинацией и поиском

- **patch** - частичное обновление

- **del** - удаление заметки
4. **Утилитарные функции:**
   
   ```go
   func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
   }
   ```
   
   Единообразная отправка JSON ответов.

**Архитектурное значение:**

- Отделяет транспортный уровень (HTTP) от бизнес-логики
- Реализует RESTful API для взаимодействия с системой
- Обрабатывает ошибки на уровне HTTP
- Обеспечивает единый формат ответов

### Файл: `cmd/api/main.go`

Точка входа приложения, инициализация зависимостей и запуск сервера.

**Ключевые компоненты:**

1. **Инициализация и конфигурация:**
   
   ```go
   uri := getenv("MONGO_URI", "mongodb://root:secret@localhost:27017/?authSource=admin")
   dbName := getenv("MONGO_DB", "pz8")
   addr := getenv("HTTP_ADDR", ":8080")
   ```
   
   Загрузка конфигурации из переменных окружения.

2. **Подключение к базе данных:**
   
   ```go
   deps, err := db.ConnectMongo(context.Background(), uri, dbName)
   if err != nil { 
    log.Fatal("mongo connect:", err) 
   }
   defer deps.Client.Disconnect(context.Background())
   ```
   
   Инициализация подключения к MongoDB с graceful shutdown.

3. **Инициализация зависимостей:**
   
   ```go
   repo, err := notes.NewRepo(deps.Database)
   h := notes.NewHandler(repo)
   ```
   
   Создание репозитория и обработчиков.

4. **Настройка маршрутов:**
   
   ```go
   r := chi.NewRouter()
   r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"status":"ok"}`))
   })
   r.Mount("/api/v1/notes", h.Routes())
   ```
   
   Добавление health-check и основных маршрутов.

5. **Graceful shutdown:**
   
   ```go
   quit := make(chan os.Signal, 1)
   signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
   <-quit
   // Завершение операций
   ```

**Архитектурное значение:**

- Единая точка входа приложения
- Инициализация всех зависимостей
- Управление жизненным циклом приложения
- Обработка сигналов завершения

---

## Документация API эндпоинтов

### Эндпоинт: `POST /notes`

Создание новой заметки.

**URL:** `http://arbond.ru/go/8/notes`

**Тело запроса:**

```json
{
    "title": "Заголовок заметки",
    "content": "Содержимое заметки"
}
```

**Успешный ответ (201 Created):**

```json
{
    "id": "651fba2f9a8b5c1234567890",
    "title": "Заголовок заметки",
    "content": "Содержимое заметки",
    "createdAt": "2025-09-27T12:00:00Z",
    "updatedAt": "2025-09-27T12:00:00Z"
}
```

**Ошибки:**

- `400 Bad Request` - неверный формат JSON или отсутствует заголовок
- `409 Conflict` - заголовок уже существует (уникальный индекс)

### Эндпоинт: `GET /notes`

Получение списка заметок с пагинацией и поиском.

**URL:** `http://arbond.ru/go/8/notes?q=поиск&limit=10&skip=0`

**Параметры:**

- `q` (опциональный) - поиск по заголовку
- `limit` (опциональный) - количество записей (по умолчанию 20, максимум 100)
- `skip` (опциональный) - пропуск записей (по умолчанию 0)

**Успешный ответ (200 OK):**

```json
[
    {
        "id": "651fba2f9a8b5c1234567890",
        "title": "Первая заметка",
        "content": "Содержимое",
        "createdAt": "2025-09-27T12:00:00Z",
        "updatedAt": "2025-09-27T12:00:00Z"
    }
]
```

### Эндпоинт: `GET /notes/{id}`

Получение заметки по идентификатору.

**URL:** `http://arbond.ru/go/8/notes/651fba2f9a8b5c1234567890`

**Успешный ответ (200 OK):**

```json
{
    "id": "651fba2f9a8b5c1234567890",
    "title": "Заголовок заметки",
    "content": "Содержимое заметки",
    "createdAt": "2025-09-27T12:00:00Z",
    "updatedAt": "2025-09-27T12:00:00Z"
}
```

**Ошибки:**

- `404 Not Found` - заметка не найдена

### Эндпоинт: `PATCH /notes/{id}`

Частичное обновление заметки.

**URL:** `http://arbond.ru/go/8/notes/651fba2f9a8b5c1234567890`

**Тело запроса:**

```json
{
    "title": "Новый заголовок",
    "content": "Обновленное содержимое"
}
```

**Успешный ответ (200 OK):**

```json
{
    "id": "651fba2f9a8b5c1234567890",
    "title": "Новый заголовок",
    "content": "Обновленное содержимое",
    "createdAt": "2025-09-27T12:00:00Z",
    "updatedAt": "2025-09-27T12:05:00Z"
}
```

**Ошибки:**

- `404 Not Found` - заметка не найдена
- `409 Conflict` - новый заголовок уже существует

### Эндпоинт: `DELETE /notes/{id}`

Удаление заметки.

**URL:** `http://arbond.ru/go/8/notes/651fba2f9a8b5c1234567890`

**Успешный ответ (204 No Content):**
Тело ответа отсутствует

**Ошибки:**

- `404 Not Found` - заметка не найдена

---

## Запуск MongoDB

### Запуск через Docker Compose (рекомендуемый)

Создайте файл `docker-compose.yml`:

```yaml
version: '3.8'
services:
  mongo:
    image: mongo:7
    container_name: mongo-dev
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: secret
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

Запуск MongoDB:

```bash
docker-compose up -d
```

Проверка статуса:

```bash
docker-compose ps
```

Подключение через mongosh:

```bash
docker exec -it mongo-dev mongosh -u root -p secret --authenticationDatabase admin
```

---

## Тестирование

### Запуск тестов

```bash
go test ./...
```

![app-test](./img/test.png)

### Создание заметки

```bash
curl -X POST http://arbond.ru/go/8/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Первая заметка","content":"Содержимое заметки"}'
```

![create](./img/create.png)

### Получение списка заметок

```bash
curl "http://arbond.ru/go/8/notes?limit=5&skip=0&q=первая"
```

![list](./img/list.png)

### Получение заметки по ID

```bash
curl "http://arbond.ru/go/8/notes/6909eec6a19a6fe81c1e8d23"
```

![note](./img/note.png)

### Частичное обновление заметки

```bash
curl -X PATCH http://arbond.ru/go/8/notes/6909eec6a19a6fe81c1e8d23 \
  -H "Content-Type: application/json" \
  -d '{"content":"Обновленное содержимое"}'
```

![update](./img/update.png)

### Удаление заметки

```bash
curl -X DELETE http://arbond.ru/go/8/notes/6909eec6a19a6fe81c1e8d23
```

![delete](./img/delete.png)

### Тестирование обработки ошибок:

#### Создание заметки без заголовка

```bash
curl -X POST http://arbond.ru/go/8/notes \
  -H "Content-Type: application/json" \
  -d '{"content":"Без заголовка"}'
```

Ответ: `400 Bad Request`

![400](./img/400.png)

#### Попытка создать дубликат заголовка

```bash
curl -X POST http://arbond.ru/go/8/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Первая заметка","content":"Дубликат"}'
```

Ответ: `409 Conflict`

![dubl](./img/dubl.png)

#### Получение несуществующей заметки

```bash
curl "http://arbond.ru/go/8/notes/507f1f77bcf86cd799439011"
```

Ответ: `404 Not Found`

![404](./img/404.png)

#### Неверный формат ID

```bash
curl "http://arbond.ru/go/8/notes/invalid-id"
```

Ответ: `404 Not Found`

![id404](./img/id404.png)

---

## Заключение

В ходе практической работы успешно реализовано Go-приложение для работы с MongoDB.
Приложение предоставляет полнофункциональное REST API для управления заметками:

- **POST /notes** - создание заметок с уникальными заголовками
- **GET /notes** - получение списка с поиском и пагинацией  
- **GET /notes/{id}** - получение заметки по идентификатору
- **PATCH /notes/{id}** - частичное обновление заметок
- **DELETE /notes/{id}** - удаление заметок

Реализованное решение демонстрирует принципы построения веб-приложений с документной базой данных
и может служить основой для более сложных систем, требующих гибкой схемы данных
и высокой производительности при работе с полуструктурированной информацией.

