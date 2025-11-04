package notes

import (
	"context"
	"os"
	"testing"

	"app/internal/db"
)

func TestCreateAndGetNote(t *testing.T) {
	ctx := context.Background()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://root:secret@localhost:27017/pz8_test?authSource=admin"
	}

	deps, err := db.ConnectMongo(ctx, uri)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer deps.Disconnect(ctx)

	// Очищаем тестовую базу
	err = deps.Database.Drop(ctx)
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	repo, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Тест создания заметки
	created, err := repo.Create(ctx, "Test Title", "Test Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if created.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", created.Title)
	}

	if created.Content != "Test Content" {
		t.Errorf("Expected content 'Test Content', got '%s'", created.Content)
	}

	if created.ID.IsZero() {
		t.Error("Expected non-zero ID")
	}

	// Тест получения заметки по ID
	found, err := repo.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to get note by ID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID.Hex(), found.ID.Hex())
	}

	if found.Title != created.Title {
		t.Errorf("Expected title '%s', got '%s'", created.Title, found.Title)
	}
}

func TestDuplicateTitle(t *testing.T) {
	ctx := context.Background()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://root:secret@localhost:27017/pz8_test_duplicate?authSource=admin"
	}

	deps, err := db.ConnectMongo(ctx, uri)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer deps.Disconnect(ctx)

	err = deps.Database.Drop(ctx)
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	repo, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Создаем первую заметку
	_, err = repo.Create(ctx, "Unique Title", "Content 1")
	if err != nil {
		t.Fatalf("Failed to create first note: %v", err)
	}

	// Пытаемся создать заметку с таким же заголовком
	_, err = repo.Create(ctx, "Unique Title", "Content 2")
	if err == nil {
		t.Error("Expected error for duplicate title, got nil")
	}

	if !IsDuplicateKeyError(err) {
		t.Errorf("Expected duplicate key error, got %v", err)
	}
}

func IsDuplicateKeyError(err error) bool {
	return err != nil && err.Error() == "duplicate title"
}

func TestUpdateNote(t *testing.T) {
	ctx := context.Background()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://root:secret@localhost:27017/pz8_test_update?authSource=admin"
	}

	deps, err := db.ConnectMongo(ctx, uri)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer deps.Disconnect(ctx)

	err = deps.Database.Drop(ctx)
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	repo, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Создаем заметку
	created, err := repo.Create(ctx, "Original Title", "Original Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Обновляем заметку
	newContent := "Updated Content"
	updated, err := repo.Update(ctx, created.ID.Hex(), nil, &newContent)
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updated.Content != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, updated.Content)
	}

	if updated.Title != created.Title {
		t.Errorf("Title should not change, expected '%s', got '%s'", created.Title, updated.Title)
	}
}
