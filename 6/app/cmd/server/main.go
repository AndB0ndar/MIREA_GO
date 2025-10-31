package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"app/internal/db"
	"app/internal/httpapi"
	"app/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	d := db.Connect()

	if err := d.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); err != nil {
		log.Fatal("migrate:", err)
	}

	r := httpapi.BuildRouter(d)

	/*
	   port := os.Getenv("PORT")
	   if port == "" { port = "8080" }
	   addr := ":" + port
	   log.Println("Server listening on", addr)
	   log.Fatal(http.ListenAndServe(addr, r))
	*/

	socketPath := os.Getenv("SOCKET_PATH")
	if socketPath == "" {
		socketPath = "/tmp/app.sock"
	}
	// Check if socket file already exists
	if _, err := os.Stat(socketPath); err == nil {
		log.Fatalf("Error: socket file already exists at %s. Remove it manually or choose a different path.", socketPath)
	}
	// Create listener on Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	// Set appropriate permissions on the socket file
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on Unix socket: ", socketPath)
	log.Fatal(http.Serve(listener, r))
}
