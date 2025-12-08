package db

import (
	"database/sql"
	"fmt"
	"log"
)

// InitSchema инициализирует схему БД
func InitSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS notes (
            id BIGSERIAL PRIMARY KEY,
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
            updated_at TIMESTAMPTZ
        )`,
		`CREATE INDEX IF NOT EXISTS idx_notes_title_gin 
         ON notes USING GIN (to_tsvector('simple', title))`,
		`CREATE INDEX IF NOT EXISTS idx_notes_created_id 
         ON notes (created_at, id)`,
		`CREATE OR REPLACE FUNCTION update_updated_at()
         RETURNS TRIGGER AS $$
         BEGIN
             NEW.updated_at = NOW();
             RETURN NEW;
         END;
         $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS trg_notes_updated ON notes`,
		`CREATE TRIGGER trg_notes_updated
         BEFORE UPDATE ON notes
         FOR EACH ROW
         EXECUTE FUNCTION update_updated_at()`,
	}

	for i, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %d: %w", i, err)
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// TruncateTables очищает таблицы для тестов
func TruncateTables(db *sql.DB) error {
	if _, err := db.Exec("TRUNCATE TABLE notes RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}
	return nil
}

// SeedTestData заполняет БД тестовыми данными
func SeedTestData(db *sql.DB) error {
	testNotes := []struct {
		title   string
		content string
	}{
		{"Test Note 1", "Content of test note 1"},
		{"Test Note 2", "Content of test note 2"},
		{"Test Note 3", "Content of test note 3"},
	}

	for _, note := range testNotes {
		_, err := db.Exec(
			"INSERT INTO notes (title, content) VALUES ($1, $2)",
			note.title, note.content,
		)
		if err != nil {
			return fmt.Errorf("failed to insert test data: %w", err)
		}
	}

	log.Println("Test data seeded successfully")
	return nil
}
