-- Создание таблицы notes
CREATE TABLE IF NOT EXISTS notes (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

-- Частичный индекс для поиска по заголовку
CREATE INDEX IF NOT EXISTS idx_notes_title_gin ON notes USING GIN (to_tsvector('simple', title));

-- Индекс для keyset-пагинации
CREATE INDEX IF NOT EXISTS idx_notes_created_id ON notes (created_at, id);

-- Расширение для статистики запросов
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Создаем функцию для генерации тестовых данных
CREATE OR REPLACE FUNCTION generate_test_notes(count integer) RETURNS void AS $$
DECLARE
    i integer;
BEGIN
    FOR i IN 1..count LOOP
        INSERT INTO notes (title, content) 
        VALUES ('Тестовая заметка ' || i, 'Содержание тестовой заметки ' || i);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Генерируем 1000 тестовых записей (можно изменить количество)
SELECT generate_test_notes(1000);

-- Удаляем функцию после использования
DROP FUNCTION generate_test_notes(integer);

-- Проверяем количество записей
SELECT 'Таблица notes создана, записей: ' || COUNT(*) FROM notes;
