.PHONY: run test bench profile load clean

run:
	go run ./cmd/api

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem -benchtime=5s ./internal/work/...

bench-cpu:
	go test -bench=. -benchmem -cpuprofile=cpu_bench.out ./internal/work/...

profile-cpu:
	curl -o cpu_profile.out "http://localhost:8080/debug/pprof/profile?seconds=30"

profile-heap:
	curl -o heap_profile.out "http://localhost:8080/debug/pprof/heap"

profile-goroutine:
	curl -o goroutine_profile.out "http://localhost:8080/debug/pprof/goroutine"

profile-block:
	curl -o block_profile.out "http://localhost:8080/debug/pprof/block"

profile-mutex:
	curl -o mutex_profile.out "http://localhost:8080/debug/pprof/mutex"

load-slow:
	hey -n 1000 -c 10 http://localhost:8080/work

load-fast:
	hey -n 10000 -c 50 http://localhost:8080/work-fast

web-cpu:
	go tool pprof -http=:9999 cpu_profile.out

web-heap:
	go tool pprof -http=:9998 heap_profile.out

web-block:
	go tool pprof -http=:9997 block_profile.out

clean:
	rm -f *.out
