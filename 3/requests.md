# Тестовые запросы для HTTP-сервера

## Проверяем, что сервер работает

```
curl http://localhost:8081/health
```

## Создаем несколько задач

```
curl -X POST http://localhost:8081/tasks -H "Content-Type: application/json" -d '{"title":"Первая задача"}'
curl -X POST http://localhost:8081/tasks -H "Content-Type: application/json" -d '{"title":"Вторая задача"}'
curl -X POST http://localhost:8081/tasks -H "Content-Type: application/json" -d '{"title":"Третья задача"}'
```

## Получаем все задачи

```
curl http://localhost:8081/tasks
```

## Обновляем первую задачу

```
curl -X PATCH http://localhost:8081/tasks/1 -H "Content-Type: application/json" -d '{"done":true}'
```

## Фильтруем задачи

```
curl "http://localhost:8081/tasks?q=Первая"
```

## Удаляем вторую задачу

```
curl -X DELETE http://localhost:8081/tasks/2
```

## Проверяем результат

```
curl http://localhost:8081/tasks
```
