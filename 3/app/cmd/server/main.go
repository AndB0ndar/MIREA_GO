package main

import (
	"log"
	"net/http"
	"os"

	"app/internal/api"
	"app/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	store := storage.NewMemoryStore()
	h := api.NewHandlers(store)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request){
		api.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /tasks", h.ListTasks)
	mux.HandleFunc("POST /tasks", h.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", h.GetTask)
	mux.HandleFunc("PATCH /tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)

	// Middleware chain: CORS -> Logging -> Mux
	handler := api.CORS(api.Logging(mux))

	log.Println("Server listening on", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
