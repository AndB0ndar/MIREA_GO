# Практическое задание 1

## Бондарь Андрей Ренатович ЭФМО-02-25

## Установка Go и Git (если не установлены)

```bash
sudo apt update
sudo apt install golang git -y
```

## Проверка версий

```bash
go version
git --version
```

## Создание структуры проекта

```bash
mkdir -p ~/helloapi/cmd/server
cd ~/helloapi
```

## Инициализация модуля

```bash
go mod init example.com/helloapi
```

## Создание main.go

```bash
vim ~/helloapi/cmd/server/main.go
```

## Установка зависимостей

```bash
go get github.com/google/uuid@latest
go mod tidy
```

## Запуск сервера

```bash
cd ~/helloapi
go run ./cmd/server
```

## Проверка в другом терминале

```bash
curl http://localhost:8080/hello
curl http://localhost:8080/user
```

## Сборка бинарника

```bash
go build -o helloapi ./cmd/server
./helloapi
```

## Проверка код-стайла

```bash
go fmt ./...
go vet ./...
```

## Настройка порта через переменные окружения

```bash
APP_PORT=9090 go run ./cmd/server
curl http://localhost:9090/hello
curl http://localhost:9090/user
```

## Остальное

Структура проекта:

```text
helloapi/
├── cmd/
│   └── server/
│       └── main.go
├── go.mod
└── go.sum
```

Все команды были протестированы на Ubuntu 22.04. Основные отличия от Windows:
1. Используются прямые слеши в путях
1. Бинарник создается без расширения .exe
1. Команды для переменных окружения отличаются (export APP_PORT=9090 вместо $env:APP_PORT="9090")

