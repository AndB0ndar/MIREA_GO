package work

import "testing"

func BenchmarkFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Fib(30)
	}
}

func BenchmarkFibFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FibFast(30)
	}
}

func BenchmarkFib40(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Fib(40)
	}
}

func BenchmarkFibFast40(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FibFast(40)
	}
}
