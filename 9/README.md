# Практическое задание 9. Реализация регистрации и входа пользователей. Хэширование паролей с bcrypt

Студент группы *ЭФМО-02-25 Бондарь Андрей Ренатович*

## Описание

**Цели:**
- Научиться безопасно хранить пароли (bcrypt), валидировать вход и обрабатывать ошибки.
- Реализовать эндпоинты POST /auth/register и POST /auth/login.
- Закрепить работу с БД (PostgreSQL + GORM или database/sql) и валидацией ввода.
- Подготовить основу для JWT-аутентификации в следующем ПЗ (No 10). 

---

## Инициализация проекта

```bash
mkdir -p pz9-auth/cmd/api pz9-auth/internal/core pz9-auth/internal/http/handlers pz9-auth/internal/repo pz9-auth/internal/platform/config
cd pz9-auth
go mod init example.com/pz9-auth
go get github.com/go-chi/chi/v5
go get gorm.io/gorm gorm.io/driver/postgres
go get golang.org/x/crypto/bcrypt

# Переменные окружения
export POSTGRESQL_DSN="postgres://user:pass@localhost:5432/pz9?sslmode=disable"
export BCRYPT_COST=12
export APP_ADDR=":8080"
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
    ├── core
    │   └── user.go
    ├── http
    │   └── handlers
    │       └── auth.go
    ├── platform
    │   └── config
    │       └── config.go
    └── repo
        ├── postgres.go
        └── user_repo.go
```

---

## Реализация

В этом разделе описывается реализация ключевых компонентов приложения с объяснением назначения каждого файла и его роли в системе.

### Файл: `internal/platform/config/config.go`

Централизованное управление конфигурацией приложения через переменные окружения.

**Ключевые компоненты:**

1. **Структура Config:**
```go
type Config struct {
    DB_DSN      string
    BcryptCost  int
    Addr        string
}
```
Объединяет все настройки приложения в одной структуре.

2. **Функция Load:**
```go
func Load() Config {
    cost := 12
    if v := os.Getenv("BCRYPT_COST"); v != "" {
        // парсинг значения
    }
    // загрузка остальных переменных
}
```
Загружает и парсит переменные окружения с установкой значений по умолчанию.

**Архитектурное значение:**
- Единая точка для управления конфигурацией
- Гибкость настройки через environment variables
- Изоляция конфигурационной логики от бизнес-кода

### Файл: `internal/core/user.go`

Определение модели данных пользователя для БД и JSON представления.

**Ключевые компоненты:**

1. **Структура User:**
```go
type User struct {
    ID           int64     `gorm:"primaryKey" json:"id"`
    Email        string    `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash string    `gorm:"not null" json:"-"`
    CreatedAt    time.Time `json:"createdAt"`
}
```
Модель с тегами GORM для работы с БД и JSON для API.

**Архитектурное значение:**
- Единая модель данных для всего приложения
- Автоматическая миграция схемы БД
- Защита конфиденциальных данных (PasswordHash исключен из JSON)

### Файл: `internal/repo/user_repo.go`

Инкапсуляция логики работы с базой данных и операций над пользователями.

**Ключевые компоненты:**

1. **Кастомные ошибки:**
```go
var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already in use")
```
Четкая семантика ошибок для разных сценариев.

2. **Метод Create:**
```go
func (r *UserRepo) Create(ctx context.Context, u *User) error {
    // проверка уникальности email
    // сохранение пользователя
}
```
Создание пользователя с обработкой конфликта email.

3. **Метод ByEmail:**
```go
func (r *UserRepo) ByEmail(ctx context.Context, email string) (User, error) {
    // поиск по email с обработкой "not found"
}
```
Поиск пользователя по email для операций входа.

**Архитектурное значение:**
- Инкапсуляция SQL/GORM логики
- Единообразная обработка ошибок
- Поддержка контекста для отмены операций

### Файл: `internal/http/handlers/auth.go`

Обработка HTTP запросов регистрации и входа с валидацией и безопасной обработкой паролей.

**Ключевые компоненты:**

1. **Структура AuthHandler:**
```go
type AuthHandler struct {
    Users      *repo.UserRepo
    BcryptCost int
}
```
Инкапсуляция зависимостей для обработчиков.

2. **Метод Register:**
```go
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // валидация входных данных
    // хэширование пароля: bcrypt.GenerateFromPassword
    // сохранение пользователя
}
```
Полный цикл регистрации с валидацией и хэшированием.

3. **Метод Login:**
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // поиск пользователя
    // сравнение паролей: bcrypt.CompareHashAndPassword
    // унифицированные сообщения об ошибках
}
```
Безопасная аутентификация с защитой от перебора.

**Архитектурное значение:**
- Отделение HTTP слоя от бизнес-логики
- Единообразные JSON ответы
- Безопасная обработка аутентификации

### Файл: `cmd/api/main.go`

Точка входа приложения, инициализация зависимостей и запуск сервера.

**Ключевые компоненты:**

1. **Инициализация зависимостей:**
```go
cfg := config.Load()
db := repo.Open(cfg.DB_DSN)
userRepo := repo.NewUserRepo(db)
authHandler := &handlers.AuthHandler{...}
```
Proper dependency injection и инициализация компонентов.

2. **Настройка маршрутов:**
```go
r := chi.NewRouter()
r.Post("/auth/register", authHandler.Register)
r.Post("/auth/login", authHandler.Login)
```
Чистая конфигурация API эндпоинтов.

**Архитектурное значение:**
- Компоновка всех компонентов приложения
- Централизованная обработка ошибок инициализации
- Чистая точка входа

---

## Документация API эндпоинтов

### Эндпоинт: `POST /auth/register`

Регистрация нового пользователя в системе.

**Полный URL:** `http://arbond.ru/go/9/auth/register`

**Параметры (JSON body):**
- `email` (обязательный) - email пользователя
- `password` (обязательный) - пароль (минимум 8 символов)

**Пример запроса:**
```http
POST http://arbond.ru/go/9/auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "Secret123!"
}
```

**Успешный ответ (201 Created):**
```json
{
    "status": "ok",
    "user": {
        "id": 1,
        "email": "user@example.com"
    }
}
```

**Ошибки:**
- `400 Bad Request` - невалидный JSON или недостаточная длина пароля
- `409 Conflict` - email уже занят
- `500 Internal Server Error` - внутренние ошибки сервера

### Эндпоинт: `POST /auth/login`

Аутентификация существующего пользователя.

**Полный URL:** `http://arbond.ru/go/9/auth/login`

**Параметры (JSON body):**
- `email` (обязательный) - email пользователя
- `password` (обязательный) - пароль пользователя

**Пример запроса:**
```http
POST http://arbond.ru/go/9/auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "Secret123!"
}
```

**Успешный ответ (200 OK):**
```json
{
    "status": "ok",
    "user": {
        "id": 1,
        "email": "user@example.com"
    }
}
```

**Ошибки:**
- `400 Bad Request` - невалидный JSON или отсутствуют параметры
- `401 Unauthorized` - неверный email или пароль (унифицированное сообщение)
- `500 Internal Server Error` - внутренние ошибки сервера

---

## Запуск PostgreSQL

### Запуск через Docker (рекомендуемый)

Запуск PostgreSQL в контейнере:
```bash
docker run --name postgres -p 5432:5432 -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=pz9 -d postgres
```

Проверка статуса:
```bash
docker ps
```

Остановка контейнера:
```bash
docker stop postgres
```

---

## Тестирование

### Регистрация пользователя

```bash
curl -X POST http://arbond.ru/go/9/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Secret123!"}'
```
Ответ: 201 Created с данными пользователя

### Повторная регистрация с тем же email
```bash
curl -X POST http://arbond.ru/go/9/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"AnotherPass"}'
```
Ответ: 409 Conflict, "email_taken"

### Успешный вход
```bash
curl -X POST http://arbond.ru/go/9/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Secret123!"}'
```
Ответ: 200 OK с данными пользователя

### Неудачный вход (неверный пароль)
```bash
curl -X POST http://arbond.ru/go/9/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrong"}'
```
Ответ: 401 Unauthorized, "invalid_credentials"

### Неудачный вход (несуществующий email)
```bash
curl -X POST http://arbond.ru/go/9/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"nonexistent@example.com","password":"any"}'
```
Ответ: 401 Unauthorized, "invalid_credentials"

### Тестирование валидации:

#### Короткий пароль
```bash
curl -X POST http://arbond.ru/go/9/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@example.com","password":"short"}'
```
Ответ: 400 Bad Request, "email_required_and_password_min_8"

#### Отсутствует email
```bash
curl -X POST http://arbond.ru/go/9/auth/register \
  -H "Content-Type: application/json" \
  -d '{"password":"Secret123!"}'
```
Ответ: 400 Bad Request, "email_required_and_password_min_8"

#### Невалидный JSON
```bash
curl -X POST http://arbond.ru/go/9/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Secret123!"'
```
Ответ: 400 Bad Request, "invalid_json"

---

## Заключение

В ходе практической работы успешно реализовано Go-приложение для безопасной аутентификации пользователей. Приложение предоставляет 2 основных эндпоинта для регистрации и входа пользователей:

- **http://arbond.ru/go/9/auth/register** - регистрация новых пользователей с хэшированием паролей
- **http://arbond.ru/go/9/auth/login** - аутентификация существующих пользователей

