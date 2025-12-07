#!/bin/bash

set -e

echo "=== Запуск нагрузочного тестирования для ПЗ 14 ==="
command -v hey >/dev/null 2>&1 || { 
    echo "Установите hey: go install github.com/rakyll/hey@latest"
    exit 1
}
command -v wrk >/dev/null 2>&1 || {
    echo "Установите wrk: sudo apt-get install wrk"
    exit 1
}


BASE_URL=${1:-http://localhost:8080}

echo "Базовый URL: $BASE_URL"
echo

echo "1. Тестирование OFFSET пагинации (неоптимизированная)..."
hey -n 1000 -c 50 "$BASE_URL/notes/offset?limit=10&offset=100"
echo

echo "2. Тестирование Keyset пагинации (оптимизированная)..."
hey -n 1000 -c 50 "$BASE_URL/notes/paginated?limit=10"
echo

echo "3. Тестирование N+1 запросов (получение 10 заметок по одной)..."
for i in {1..10}; do
    time curl -s -o /dev/null "$BASE_URL/notes/$i" &
done
wait
echo

echo "4. Тестирование батчинга (решение N+1)..."
hey -n 1000 -c 50 "$BASE_URL/notes/batch?ids=1,2,3,4,5,6,7,8,9,10"
echo

echo "5. Тестирование поиска по заголовку..."
hey -n 500 -c 20 "$BASE_URL/notes/search?q=заметка&limit=10"
echo

echo "6. Проверка статистики БД..."
curl -s "$BASE_URL/api/stats/db" | jq .
echo

echo "7. Проверка EXPLAIN запроса..."
curl -s "$BASE_URL/notes/explain?query=SELECT%20*%20FROM%20notes%20WHERE%20id%20=%201" | jq -r .explanation
echo

echo "=== Нагрузочное тестирование завершено ==="
