# Практическое задание 11. Проектирование REST API (CRUD для заметок)

Студент группы *ЭФМО-02-25 Бондарь Андрей Ренатович*

## Описание

**Цели:**
- Освоить принципы проектирования REST API.
- Научиться разрабатывать структуру проекта backend-приложения на Go.
- Спроектировать и реализовать CRUD-интерфейс для сущности "Заметка".
- Освоить применение слоистой архитектуры (handler -> service -> repository).
- Подготовить основу для интеграции с базой данных и JWT-аутентификацией.

---

## Инициализация проекта

```bash
mkdir -p app/{cmd/{api,server},internal/{http/handlers,core,repo,platform/{config}},api}
cd app
go mod init app
go get github.com/go-chi/chi/v5
```

---

## Структура проекта

```
app
├── cmd
│   ├── api
│   │   └── main.go
│   └── server
│       └── main.go
├── go.mod
├── go.sum
└── internal
    ├── core
    │   └── note.go
    ├── http
    │   ├── handlers
    │   │   └── notes.go
    │   └── router.go
    └── repo
        └── note_mem.go
```

---

## Реализация

В этом разделе описывается реализация ключевых компонентов приложения с объяснением назначения каждого файла и его роли в системе.

### Файл: `internal/core/note.go`

Определение моделей данных и бизнес-сущностей.

**Ключевые компоненты:**

1. **Структура Note:**
```go
type Note struct {
    ID        int64      `json:"id"`
    Title     string     `json:"title"`
    Content   string     `json:"content"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}
```
Основная сущность приложения с JSON-тегами для сериализации.

2. **Структура NoteRequest:**
```go
type NoteRequest struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}
```
DTO для входящих запросов на создание/обновление заметок.

**Архитектурное значение:**
- Определяет доменную модель приложения
- Отделяет структуры данных от транспортного уровня
- Обеспечивает согласованность данных между слоями

### Файл: `internal/repo/note_mem.go`

Реализация слоя доступа к данным с in-memory хранилищем.

**Ключевые компоненты:**

1. **Интерфейс NoteRepository:**
```go
type NoteRepository interface {
    Create(note Note) (int64, error)
    GetByID(id int64) (*Note, error)
    GetAll() ([]*Note, error)
    Update(id int64, updated Note) error
    Delete(id int64) error
}
```
Абстракция для работы с данными, позволяющая легко менять реализацию хранилища.

2. **In-memory реализация:**
```go
type NoteRepoMem struct {
    mu     sync.RWMutex
    notes  map[int64]*Note
    nextID int64
}
```
Потокобезопасная реализация с использованием sync.RWMutex для конкурентного доступа.

3. **CRUD операции:**
- `Create()` - создание новой заметки с автоинкрементом ID
- `GetByID()` - получение заметки по идентификатору
- `GetAll()` - получение всех заметок
- `Update()` - обновление существующей заметки
- `Delete()` - удаление заметки

**Архитектурное значение:**
- Инкапсулирует логику работы с данными
- Предоставляет абстракцию для легкой замены хранилища
- Обеспечивает потокобезопасность

### Файл: `internal/http/handlers/notes.go`

Обработчики HTTP-запросов, реализующие REST API.

**Ключевые компоненты:**

1. **Структура Handler:**
```go
type Handler struct {
    Repo NoteRepository
}
```
Инкапсулирует зависимости обработчиков.

2. **CRUD эндпоинты:**
- `CreateNote()` - POST /api/v1/notes
- `GetAllNotes()` - GET /api/v1/notes  
- `GetNote()` - GET /api/v1/notes/{id}
- `UpdateNote()` - PATCH /api/v1/notes/{id}
- `DeleteNote()` - DELETE /api/v1/notes/{id}

3. **Обработка ошибок:**
Каждый обработчик корректно обрабатывает различные сценарии ошибок с соответствующими HTTP-статусами.

**Архитектурное значение:**
- Отделяет транспортный уровень от бизнес-логики
- Обрабатывает HTTP-специфичные аспекты (коды ответов, заголовки)
- Валидирует входящие данные

### Файл: `internal/http/router.go`

Конфигурация маршрутизации API.

**Ключевые компоненты:**

```go
func NewRouter(h *handlers.Handler) *chi.Mux {
    r := chi.NewRouter()
    
    r.Route("/api/v1/notes", func(r chi.Router) {
        r.Post("/", h.CreateNote)
        r.Get("/", h.GetAllNotes)
        r.Get("/{id}", h.GetNote)
        r.Patch("/{id}", h.UpdateNote)
        r.Delete("/{id}", h.DeleteNote)
    })
    
    return r
}
```

**Архитектурное значение:**
- Централизованная конфигурация маршрутов
- Чистое разделение ответственности

### Файл: `cmd/api/main.go`

Точка входа приложения, инициализация зависимостей.

**Ключевые компоненты:**

```go
noteRepo := repo.NewNoteRepoMem()
handler := &handlers.Handler{Repo: noteRepo}
router := httpx.NewRouter(handler)
```

**Архитектурное значение:**
- Композиция зависимостей (Dependency Injection)
- Централизованная точка запуска приложения
- Обработка ошибок запуска

---

## Документация API эндпоинтов

### Эндпоинт: [POST /notes](http://arbond.ru/go/11/notes)

Создание новой заметки.

**URL:** `http://arbond.ru/go/11/notes`

**Тело запроса (JSON):**
```json
{
    "title": "Заголовок заметки",
    "content": "Содержание заметки"
}
```

**Успешный ответ (201 Created):**
```json
{
    "id": 1,
    "title": "Заголовок заметки",
    "content": "Содержание заметки",
    "createdAt": "2025-01-15T10:30:00Z",
    "updatedAt": null
}
```

**Ошибки:**
- `400 Bad Request` - некорректные входные данные
- `500 Internal Server Error` - внутренняя ошибка сервера

### Эндпоинт: [GET /notes](http://arbond.ru/go/11/notes)

Получение списка всех заметок.

**URL:** `http://arbond.ru/go/11/notes`

**Успешный ответ (200 OK):**
```json
[
    {
        "id": 1,
        "title": "Первая заметка",
        "content": "Содержание",
        "createdAt": "2025-01-15T10:30:00Z",
        "updatedAt": null
    }
]
```

### Эндпоинт: [GET /notes/{id}](http://arbond.ru/go/11/notes/{id})

Получение заметки по ID.

**URL:** `http://arbond.ru/go/11/notes/1`

**Успешный ответ (200 OK):**
```json
{
    "id": 1,
    "title": "Заметка",
    "content": "Содержание",
    "createdAt": "2025-01-15T10:30:00Z",
    "updatedAt": null
}
```

**Ошибки:**
- `404 Not Found` - заметка не найдена

### Эндпоинт: [PATCH /notes/{id}](http://arbond.ru/go/11/notes/{id})

Обновление существующей заметки.

**URL:** `http://arbond.ru/go/11/notes/1`

**Тело запроса (JSON):**
```json
{
    "title": "Обновленный заголовок",
    "content": "Обновленное содержание"
}
```

**Успешный ответ (200 OK):**
```json
{
    "id": 1,
    "title": "Обновленный заголовок",
    "content": "Обновленное содержание",
    "createdAt": "2025-01-15T10:30:00Z",
    "updatedAt": "2025-01-15T11:00:00Z"
}
```

**Ошибки:**
- `404 Not Found` - заметка не найдена

### Эндпоинт: [DELETE /notes/{id}](http://arbond.ru/go/11/notes/{id})

Удаление заметки.

**URL:** `http://arbond.ru/go/11/notes/1`

**Успешный ответ:** `204 No Content`

**Ошибки:**
- `404 Not Found` - заметка не найдена

---

## Тестирование

### Создание заметки

```bash
curl -X POST http://arbond.ru/go/11/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Первая заметка","content":"Это тестовая заметка"}'
```

**Ответ:**
```json
{
  "id": 1,
  "title": "Первая заметка",
  "content": "Это тестовая заметка",
  "createdAt": "2025-01-15T10:30:45.123Z",
  "updatedAt": null
}
```

### Получение всех заметок

```bash
curl -X GET http://arbond.ru/go/11/notes
```

**Ответ:**
```json
[
  {
    "id": 1,
    "title": "Первая заметка",
    "content": "Это тестовая заметка",
    "createdAt": "2025-01-15T10:30:45.123Z",
    "updatedAt": null
  }
]
```

### Получение конкретной заметки

```bash
curl -X GET http://arbond.ru/go/11/notes/1
```

**Ответ:**
```json
{
  "id": 1,
  "title": "Первая заметка",
  "content": "Это тестовая заметка",
  "createdAt": "2025-01-15T10:30:45.123Z",
  "updatedAt": null
}
```

### Обновление заметки

```bash
curl -X PATCH http://arbond.ru/go/11/notes/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Обновленный заголовок","content":"Обновленное содержание"}'
```

**Ответ:**
```json
{
  "id": 1,
  "title": "Обновленный заголовок",
  "content": "Обновленное содержание",
  "createdAt": "2025-01-15T10:30:45.123Z",
  "updatedAt": "2025-01-15T10:35:22.456Z"
}
```

### Удаление заметки

```bash
curl -X DELETE http://arbond.ru/go/11/notes/1
```

**Ответ:** `204 No Content`

### Тестирование ошибок

#### Создание заметки без заголовка

```bash
curl -X POST http://arbond.ru/go/11/notes \
  -H "Content-Type: application/json" \
  -d '{"content":"Заметка без заголовка"}'
```

**Ответ:** `400 Bad Request`

#### Получение несуществующей заметки

```bash
curl -X GET http://arbond.ru/go/11/notes/999
```

**Ответ:** `404 Not Found`

---

## Заключение

В ходе практической работы успешно реализовано REST API
для управления заметками с полным CRUD-функционалом. Основные достижения:

- **Спроектирована чистая архитектура** с четким разделением на слои (handler -> service -> repository)
- **Реализованы все CRUD операции** для сущности "Заметка"
- **Обеспечена корректная обработка ошибок** с соответствующими HTTP-статусами
- **Реализовано потокобезопасное in-memory хранилище**
- **Подготовлена основа** для дальнейшего расширения (база данных, аутентификация)

