package main

import (
	"fmt"
	"log"
	"net/http"
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

	log.Println("Server on ", cfg.Port, "; pprof on /debug/pprof/")
	log.Fatal(http.ListenAndServe(cfg.Port, nil))
}
