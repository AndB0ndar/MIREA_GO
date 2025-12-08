package stringsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClip(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"EmptyString_Max5", "", 5, ""},
		{"EmptyString_Max0", "", 0, ""},
		{"EmptyString_MaxNegative", "", -1, ""},
		{"Hello_Max0", "hello", 0, ""},
		{"Hello_Max3", "hello", 3, "hel"},
		{"Hello_Max5", "hello", 5, "hello"},
		{"Hello_Max10", "hello", 10, "hello"},
		{"Test_MaxNegative", "test", -5, ""},
		{"SingleChar_Max1", "a", 1, "a"},
		{"SingleChar_Max0", "a", 0, ""},
		{"LongString_Max10", "this is a long string", 10, "this is a "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Clip(tc.input, tc.max)
			assert.Equal(t, tc.expected, result,
				"Clip(%q, %d) = %q, expected %q",
				tc.input, tc.max, result, tc.expected)
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"привет", "тевирп"},
		{"a", "a"},
		{"", ""},
		{"hello world", "dlrow olleh"},
		{"12345", "54321"},
		{"😀👍", "👍😀"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Reverse(tt.input)
			assert.Equal(t, tt.expected, result)

			assert.Equal(t, tt.input, Reverse(result))
		})
	}
}

func TestContainsAny(t *testing.T) {
	t.Run("содержит одну из подстрок", func(t *testing.T) {
		result := ContainsAny("hello world", "world", "test")
		assert.True(t, result)
	})

	t.Run("не содержит ни одной", func(t *testing.T) {
		result := ContainsAny("hello world", "foo", "bar")
		assert.False(t, result)
	})

	t.Run("пустые подстроки", func(t *testing.T) {
		result := ContainsAny("hello", "")
		assert.True(t, result)
	})

	t.Run("без подстрок", func(t *testing.T) {
		result := ContainsAny("hello")
		assert.False(t, result)
	})
}

func TestPanic(t *testing.T) {
	panicFunc := func(s string) {
		if s == "" {
			panic("пустая строка")
		}
	}

	t.Run("паника при пустой строке", func(t *testing.T) {
		assert.Panics(t, func() {
			panicFunc("")
		})
	})

	t.Run("нет паники", func(t *testing.T) {
		assert.NotPanics(t, func() {
			panicFunc("hello")
		})
	})

	t.Run("паника с конкретным сообщением", func(t *testing.T) {
		assert.PanicsWithValue(t, "пустая строка", func() {
			panicFunc("")
		})
	})
}
