# 02 — Тестування

## Структура тесту

```go
func TestFunctionName(t *testing.T) {
    // arrange
    input := 42
    // act
    result := MyFunc(input)
    // assert
    if result != expected {
        t.Errorf("got %d, want %d", result, expected)
    }
}
```

## Table-driven tests (найважливіший патерн)

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

## t.Run — subtests

```go
t.Run("group/subtest", func(t *testing.T) { ... })

// Паралельне виконання
t.Run("parallel", func(t *testing.T) {
    t.Parallel() // цей subtest запускається паралельно з іншими
    // ...
})
```

## t.Helper()

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper() // рядок помилки вказує на caller, не на цю функцію
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

## testify

```go
import (
    "github.com/stretchr/testify/assert"  // не зупиняє тест при failure
    "github.com/stretchr/testify/require" // зупиняє тест (t.FailNow)
)

assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.ErrorIs(t, err, ErrNotFound)
assert.Contains(t, slice, item)

require.NoError(t, err)      // якщо err != nil — зупиняємо тест
require.NotNil(t, result)    // подальший код вимагає non-nil
```

## httptest

```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "Alice")
}
```

## Benchmarks

```go
func BenchmarkAdd(b *testing.B) {
    for b.Loop() {  // Go 1.24+: b.Loop() замість range b.N
        Add(1, 2)
    }
}
```

```bash
go test -bench=. -benchmem ./...
# BenchmarkAdd-8   1000000000   0.2 ns/op   0 B/op   0 allocs/op
```

## Fuzz testing (Go 1.18+)

```go
func FuzzReverse(f *testing.F) {
    f.Add("hello")  // seed corpus
    f.Fuzz(func(t *testing.T, s string) {
        reversed := Reverse(s)
        if Reverse(reversed) != s {
            t.Errorf("double reverse failed for %q", s)
        }
    })
}
```

```bash
go test -fuzz=FuzzReverse -fuzztime=30s
```

## Корисні прапорці

```bash
go test ./...                    # всі пакети
go test -v ./...                 # verbose
go test -run TestAdd ./...       # фільтр по імені
go test -run TestAdd/positive    # конкретний subtest
go test -count=1 ./...           # відключити кешування результатів
go test -race ./...              # детектор race condition
go test -cover ./...             # покриття
go test -coverprofile=c.out && go tool cover -html=c.out
```
