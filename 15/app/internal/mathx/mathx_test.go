package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSum_Table(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"положительные числа", 2, 3, 5},
		{"отрицательные числа", -2, -3, -5},
		{"положительное и отрицательное", 10, -5, 5},
		{"ноль", 0, 0, 0},
		{"ноль и положительное", 0, 7, 7},
		{"ноль и отрицательное", 0, -7, -7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sum(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Sum(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestSum_Assert(t *testing.T) {
	assert.Equal(t, 5, Sum(2, 3))
	assert.Equal(t, -5, Sum(-2, -3))
	assert.Equal(t, 0, Sum(0, 0))
}

func TestDivide_OkAndError(t *testing.T) {
	t.Run("Успешное деление", func(t *testing.T) {
		result, err := Divide(10, 2)
		require.NoError(t, err) // Если ошибка, тест останавливается
		assert.Equal(t, 5, result)
	})

	t.Run("Деление на ноль", func(t *testing.T) {
		result, err := Divide(10, 0)
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.EqualError(t, err, "divide by zero")
	})
}

func TestDivide_Table(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr bool
		errMsg  string
	}{
		{"обычное деление", 6, 3, 2, false, ""},
		{"деление на 1", 10, 1, 10, false, ""},
		{"деление на ноль", 5, 0, 0, true, "divide by zero"},
		{"деление нуля", 0, 5, 0, false, ""},
		{"отрицательное деление", -10, 2, -5, false, ""},
		{"деление отрицательного", 10, -2, -5, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMax - тесты для функции Max
func TestMax(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{-1, -2, -1},
		{0, 0, 0},
		{-5, 5, 5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := Max(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sum(123, 456)
	}
}

func BenchmarkDivide(b *testing.B) {
	b.Run("без ошибок", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Divide(100, 2)
		}
	})

	b.Run("с проверкой ошибки", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Divide(100, 0)
			_ = err
		}
	})
}
