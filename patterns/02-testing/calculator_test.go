package calculator_test

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	calc "go-from-to/patterns/02-testing"
)

// ================================================================
// БАЗОВІ ТЕСТИ
// ================================================================

func TestAdd(t *testing.T) {
	got := calc.Add(2, 3)
	if got != 5 {
		t.Errorf("Add(2,3) = %f, want 5", got)
	}
}

// ================================================================
// TABLE-DRIVEN TESTS
// ================================================================

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr error
	}{
		{"positive", 10, 2, 5, nil},
		{"negative dividend", -10, 2, -5, nil},
		{"float result", 7, 2, 3.5, nil},
		{"division by zero", 5, 0, 0, calc.ErrDivisionByZero},
		{"zero divided by zero", 0, 0, 0, calc.ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Divide(tt.a, tt.b)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		want    float64
		wantErr error
	}{
		{"perfect square", 9, 3, nil},
		{"float result", 2, math.Sqrt(2), nil},
		{"zero", 0, 0, nil},
		{"negative", -1, 0, calc.ErrNegativeSqrt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Sqrt(tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestCompute(t *testing.T) {
	tests := []struct {
		name    string
		input   []float64
		want    calc.Stats
		wantErr bool
	}{
		{
			name:  "basic",
			input: []float64{1, 2, 3, 4, 5},
			want:  calc.Stats{Min: 1, Max: 5, Sum: 15, Count: 5, Avg: 3},
		},
		{
			name:  "single element",
			input: []float64{42},
			want:  calc.Stats{Min: 42, Max: 42, Sum: 42, Count: 1, Avg: 42},
		},
		{
			name:  "negative values",
			input: []float64{-3, -1, -2},
			want:  calc.Stats{Min: -3, Max: -1, Sum: -6, Count: 3, Avg: -2},
		},
		{
			name:    "empty input",
			input:   []float64{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Compute(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Count, got.Count)
			assert.InDelta(t, tt.want.Min, got.Min, 1e-9)
			assert.InDelta(t, tt.want.Max, got.Max, 1e-9)
			assert.InDelta(t, tt.want.Sum, got.Sum, 1e-9)
			assert.InDelta(t, tt.want.Avg, got.Avg, 1e-9)
		})
	}
}

// ================================================================
// SUBTESTS + ПАРАЛЕЛЬНЕ ВИКОНАННЯ
// ================================================================

func TestArithmetic(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 5.0, calc.Add(2, 3))
		assert.Equal(t, 0.0, calc.Add(-1, 1))
	})

	t.Run("Subtract", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 1.0, calc.Subtract(3, 2))
	})

	t.Run("Multiply", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 6.0, calc.Multiply(2, 3))
	})
}

// ================================================================
// HELPER ФУНКЦІЯ
// ================================================================

func assertStats(t *testing.T, got, want calc.Stats) {
	t.Helper() // рядок помилки вказує на caller
	assert.Equal(t, want.Count, got.Count, "Count")
	assert.InDelta(t, want.Min, got.Min, 1e-9, "Min")
	assert.InDelta(t, want.Max, got.Max, 1e-9, "Max")
	assert.InDelta(t, want.Avg, got.Avg, 1e-9, "Avg")
}

func TestComputeWithHelper(t *testing.T) {
	got, err := calc.Compute([]float64{1, 5, 3})
	require.NoError(t, err)
	assertStats(t, got, calc.Stats{Min: 1, Max: 5, Count: 3, Avg: 3})
}

// ================================================================
// TESTIFY: assert vs require
// ================================================================

func TestTestify(t *testing.T) {
	result, err := calc.Divide(10, 2)

	// require зупиняє тест якщо провалюється (FailNow)
	require.NoError(t, err, "Divide не мала повернути помилку")
	require.Equal(t, 5.0, result)

	// assert продовжує тест (накопичує failures)
	assert.InDelta(t, 5.0, result, 0.001)
	assert.NotZero(t, result)

	// errors.Is через testify
	_, err = calc.Divide(1, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, calc.ErrDivisionByZero))
	// або коротко:
	assert.ErrorIs(t, err, calc.ErrDivisionByZero)
}

// ================================================================
// HTTPTEST
// ================================================================

func TestHTTPHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": 42}`))
	})

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{"GET success", http.MethodGet, http.StatusOK, `"result": 42`},
		{"POST not allowed", http.MethodPost, http.StatusMethodNotAllowed, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/calculate", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ================================================================
// BENCHMARKS
// ================================================================

func BenchmarkAdd(b *testing.B) {
	for b.Loop() {
		calc.Add(1.5, 2.5)
	}
}

func BenchmarkDivide(b *testing.B) {
	for b.Loop() {
		calc.Divide(10, 3) //nolint:errcheck
	}
}

func BenchmarkFactorial(b *testing.B) {
	for b.Loop() {
		calc.Factorial(20)
	}
}

func BenchmarkReverse(b *testing.B) {
	s := "Hello, Gophers! 🚀"
	for b.Loop() {
		calc.Reverse(s)
	}
}

// Benchmark з різними розмірами вводу
func BenchmarkCompute(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		nums := make([]float64, size)
		for i := range nums {
			nums[i] = float64(i)
		}
		b.Run("", func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				calc.Compute(nums) //nolint:errcheck
			}
		})
	}
}

// ================================================================
// FUZZ TESTING (Go 1.18+)
// ================================================================

func FuzzReverse(f *testing.F) {
	// Seed corpus — базові приклади
	f.Add("")
	f.Add("a")
	f.Add("hello")
	f.Add("Привіт")
	f.Add("Go🚀")

	f.Fuzz(func(t *testing.T, s string) {
		// Інваріант: подвійний reverse = оригінал
		reversed := calc.Reverse(s)
		doubleReversed := calc.Reverse(reversed)
		if doubleReversed != s {
			t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, doubleReversed, s)
		}
	})
}
