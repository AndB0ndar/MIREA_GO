package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"

	_ "net/http/pprof"

	"app/internal/platform/config"
	"app/internal/work"
)

func enableProfiling() {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
}

func main() {
	cfg := config.Load()

	enableProfiling()

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		defer work.TimeIt("Fib(38)")()
		res := work.Fib(38)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(fmt.Sprintf("%d\n", res)))
	})

	http.HandleFunc("/work-fast", func(w http.ResponseWriter, r *http.Request) {
		defer work.TimeIt("FibFast(38)")()
		res := work.FibFast(38)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(fmt.Sprintf("%d\n", res)))
	})

	if _, err := os.Stat(cfg.SocketPath); err == nil {
		log.Fatal("Error: socket file already exists at ", cfg.SocketPath)
	}
	listener, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		log.Fatal("Failed to create Unix socket listener:", err)
	}
	defer func() {
		listener.Close()
		os.Remove(cfg.SocketPath)
	}()
	if err := os.Chmod(cfg.SocketPath, 0666); err != nil {
		log.Fatal("Failed to set socket permissions:", err)
	}

	log.Println("Server starting on Unix socket:", cfg.SocketPath)
	log.Fatal(http.Serve(listener, nil))
}
