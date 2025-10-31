package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"app/internal/cache"
)

func main() {
	_ = godotenv.Load()

	redis_host := os.Getenv("REDIS_HOST")
	c := cache.New(redis_host)
	defer c.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, "key and value required", http.StatusBadRequest)
			return
		}
		err := c.Set(key, value, 10*time.Second) // TTL = 10 секунд
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "OK: %s=%s (TTL 10s)", key, value)
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		val, err := c.Get(key)
		if err != nil {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "VALUE: %s=%s", key, val)
	})
	mux.HandleFunc("/ttl", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		ttl, err := c.TTL(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "TTL for %s: %v", key, ttl)
	})
	mux.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "DELETE endpoint would remove key: %s", key)
	})

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
	log.Println("Available endpoints:")
	log.Println("  GET /set?key=name&value=John")
	log.Println("  GET /get?key=name")
	log.Println("  GET /ttl?key=name")
	log.Println("  GET /del?key=name")

	log.Fatal(http.Serve(listener, mux))
}
