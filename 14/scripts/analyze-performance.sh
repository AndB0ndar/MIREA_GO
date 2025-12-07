#!/bin/bash

echo "=== Быстрый анализ производительности БД ==="

echo "1. Проверка подключения к БД..."
docker-compose exec postgres pg_isready -U user -d notes

echo -e "\n2. Топ 10 медленных запросов:"
docker-compose exec -T postgres psql -U user -d notes -c "
SELECT 
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    rows,
    100.0 * total_exec_time / sum(total_exec_time) OVER() as percentage
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;"

echo -e "\n3. Индексы в базе:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;"

echo -e "\n4. Статистика таблицы notes:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    schemaname,
    relname as tablename,
    n_live_tup as rows_count,
    n_dead_tup as dead_rows,
    last_autovacuum,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE relname = 'notes';"

echo -e "\n5. Активные соединения:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    COUNT(*)::text as total_connections,
    COUNT(*) FILTER (WHERE state = 'active')::text as active_connections,
    COUNT(*) FILTER (WHERE state = 'idle')::text as idle_connections
FROM pg_stat_activity;"

echo -e "\n6. Размер БД и таблиц:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    'Database' as name,
    pg_size_pretty(pg_database_size('notes')) as size
UNION ALL
SELECT 
    'Table notes' as name,
    pg_size_pretty(pg_total_relation_size('public.notes')) as size;"

echo -e "\n7. Проверка настроек пула соединений:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT name, setting, unit, context FROM pg_settings 
WHERE name IN ('max_connections', 'shared_buffers', 'work_mem', 'maintenance_work_mem')
UNION ALL
SELECT 'current_connections', COUNT(*)::text, 'connections', 'user' FROM pg_stat_activity;"

echo -e "\n8. Статистика использования индексов:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    schemaname,
    relname as tablename,
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;"

echo -e "\n9. Размеры всех таблиц и индексов:"
docker-compose exec postgres psql -U user -d notes -c "
SELECT 
    schemaname || '.' || tablename as table_name,
    pg_size_pretty(pg_total_relation_size(schemaname || '.' || tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname || '.' || tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname || '.' || tablename) - 
                   pg_relation_size(schemaname || '.' || tablename)) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname || '.' || tablename) DESC;"
