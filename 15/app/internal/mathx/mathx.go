package mathx

import "errors"

func Sum(a, b int) int {
	return a + b
}

func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
