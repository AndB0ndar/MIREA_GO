package main

import (
    "log"
    "net/http"
    "os"

    "app/internal/db"
    "app/internal/httpapi"
    "app/internal/models"

    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    addr := ":" + port

    d := db.Connect()

    if err := d.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); err != nil {
        log.Fatal("migrate:", err)
    }

    r := httpapi.BuildRouter(d)
    log.Println("Server listening on", addr)
    log.Fatal(http.ListenAndServe(addr, r))
}
