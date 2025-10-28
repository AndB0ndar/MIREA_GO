package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "app/internal/cache"
)

func main() {
    // Инициализация кэша
    c := cache.New("localhost:6379")
    defer c.Close()

    mux := http.NewServeMux()

    // Эндпоинт для установки значения
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

    // Эндпоинт для получения значения
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

    // Эндпоинт для проверки TTL
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

    // Эндпоинт для удаления ключа
    mux.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        if key == "" {
            http.Error(w, "key required", http.StatusBadRequest)
            return
        }

        // Для удаления добавим метод Delete в cache.go
        // В реальном приложении нужно добавить метод Delete в структуру Cache
        fmt.Fprintf(w, "DELETE endpoint would remove key: %s", key)
    })


    _ = godotenv.Load()

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    addr := ":" + port

    prefix := GetAPIPrefix()
    if prefix != "" {
        mainMux := http.NewServeMux()
        mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Welcome to the homepage!")
        })
        mainMux.Handle(prefix + "/", http.StripPrefix(prefix, mux))
    }

    log.Println("Server listening on ", addr)
    log.Println("Available endpoints:")
    log.Println("  GET /set?key=name&value=John")
    log.Println("  GET /get?key=name")
    log.Println("  GET /ttl?key=name")
    log.Println("  GET /del?key=name")
    log.Fatal(http.ListenAndServe(addr, mainMux))
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
