package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"app/internal/db"
	"app/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf("Not found variable DATABASE_URL!")
	}

	database, err := db.OpenDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer database.Close()

	repo := repository.NewRepo(database)

	// 1) Вставка нескольких задач
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask error: %v", err)
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	// 2) Массовая вставка через транзакцию
	bulkTitles := []string{"Массовая задача 1", "Массовая задача 2", "Массовая задача 3"}
	if err := repo.CreateMany(ctx, bulkTitles); err != nil {
		log.Printf("CreateMany error: %v", err)
	} else {
		log.Println("Mass insertion completed successfully")
	}

	// 3) Обновление статуса задачи
	if err := repo.UpdateTaskStatus(ctx, 1, true); err != nil {
		log.Printf("UpdateTaskStatus error: %v", err)
	} else {
		log.Println("Task status updated")
	}

	// 4) Получение списка всех задач
	ctxList, cancelList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelList()

	tasks, err := repo.ListTasks(ctxList)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}

	fmt.Println("\n=== All Tasks ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 5) Получение выполненных задач
	doneTasks, err := repo.ListDone(ctxList, true)
	if err != nil {
		log.Printf("ListDone error: %v", err)
	} else {
		fmt.Println("\n=== Done Tasks ===")
		for _, t := range doneTasks {
			fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
				t.ID, t.Title, t.Done, t.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// 6) Поиск задачи по ID
	task, err := repo.FindByID(ctxList, 1)
	if err != nil {
		log.Printf("FindByID error: %v", err)
	} else {
		fmt.Printf("\n=== Task #1 ===\n")
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			task.ID, task.Title, task.Done, task.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
