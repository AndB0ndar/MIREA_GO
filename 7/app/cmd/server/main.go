package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/joho/godotenv"

    "app/internal/cache"
)

func main() {
    c := cache.New("localhost:6379")
    defer c.Close()

    mux := http.NewServeMux()

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

    _ = godotenv.Load()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    addr := ":" + port

    // Get prefix from environment and build the router
    prefix := GetAPIPrefix()
    mainMux := BuildAPIPrefix(prefix, mux)

    log.Println("Server listening on ", addr)
    log.Println("API Prefix:", prefix)
    log.Println("Available endpoints:")
    
    // Update the logged endpoints to show the actual prefixed URLs
    basePath := prefix
    if basePath == "" {
        basePath = "/"
    } else {
        basePath = basePath + "/"
    }

    log.Println("Server listening on ", addr)
    log.Println("Available endpoints:")
    log.Println("  GET /set?key=name&value=John")
    log.Println("  GET /get?key=name")
    log.Println("  GET /ttl?key=name")
    log.Println("  GET /del?key=name")
    log.Fatal(http.ListenAndServe(addr, mainMux))
}

func BuildAPIPrefix(prefix string, originalMux http.Handler) http.Handler {
    if prefix == "" {
        return originalMux
    }
    if !strings.HasPrefix(prefix, "/") {
        prefix = "/" + prefix
    }
    prefixedMux := http.NewServeMux()
    prefixedMux.Handle(prefix+"/", http.StripPrefix(prefix, originalMux))
    return prefixedMux
}

func GetAPIPrefix() string {
    prefix := os.Getenv("API_PREFIX")
    prefix = strings.TrimSpace(prefix)
    if prefix != "" && !strings.HasPrefix(prefix, "/") {
        prefix = "/" + prefix
    }
    prefix = strings.TrimSuffix(prefix, "/")
    return prefix
}

