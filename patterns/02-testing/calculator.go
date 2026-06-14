package calculator

import (
	"errors"
	"math"
)

var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrNegativeSqrt   = errors.New("square root of negative number")
)

func Add(a, b float64) float64      { return a + b }
func Subtract(a, b float64) float64 { return a - b }
func Multiply(a, b float64) float64 { return a * b }

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, ErrNegativeSqrt
	}
	return math.Sqrt(n), nil
}

// Factorial — рекурсивна, для fuzz тесту
func Factorial(n int) int {
	if n < 0 {
		return -1
	}
	if n == 0 {
		return 1
	}
	return n * Factorial(n-1)
}

// Reverse — для fuzz тесту
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Stats — для table-driven тестів з кількома виходами
type Stats struct {
	Min, Max, Sum float64
	Count         int
	Avg           float64
}

func Compute(nums []float64) (Stats, error) {
	if len(nums) == 0 {
		return Stats{}, errors.New("empty input")
	}
	s := Stats{Min: nums[0], Max: nums[0], Count: len(nums)}
	for _, n := range nums {
		s.Sum += n
		if n < s.Min {
			s.Min = n
		}
		if n > s.Max {
			s.Max = n
		}
	}
	s.Avg = s.Sum / float64(s.Count)
	return s, nil
}
