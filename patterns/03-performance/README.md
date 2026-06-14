# 03 — Продуктивність

## Profiling з pprof

```bash
# CPU профіль (30 секунд)
go test -cpuprofile=cpu.out -bench=. ./...
go tool pprof cpu.out

# Memory профіль
go test -memprofile=mem.out -bench=. ./...
go tool pprof mem.out

# У pprof interactive shell:
(pprof) top10        # топ функцій
(pprof) list MyFunc  # рядки конкретної функції
(pprof) web          # відкрити граф у браузері
```

### HTTP pprof endpoint (для запущеного сервісу)

```go
import _ "net/http/pprof"  // реєструє /debug/pprof/*

go http.ListenAndServe(":6060", nil)
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Escape analysis

Визначає чи змінна йде в heap (=allocation) чи залишається в stack:

```bash
go build -gcflags='-m' ./...
# ./main.go:10:6: moved to heap: x   ← allocation
# ./main.go:15:10: ... does not escape ← no allocation
```

## sync.Pool

Повторне використання об'єктів щоб уникнути GC pressure:

```go
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    buf.Write(data)
    return buf.String()
}
```

## Slice pre-allocation

```go
// Погано: багато реалокацій
var result []int
for i := range 1000 {
    result = append(result, i*2)
}

// Добре: одна алокація
result := make([]int, 0, 1000)
for i := range 1000 {
    result = append(result, i*2)
}

// Або якщо розмір точно відомий
result := make([]int, 1000)
for i := range 1000 {
    result[i] = i * 2
}
```

## strings.Builder vs конкатенація

```go
// Погано: O(n²) — кожен += створює новий рядок
s := ""
for i := range 1000 {
    s += fmt.Sprintf("item%d", i)
}

// Добре: O(n) — один буфер
var sb strings.Builder
sb.Grow(1000 * 7) // опціонально: pre-alloc
for i := range 1000 {
    fmt.Fprintf(&sb, "item%d", i)
}
s := sb.String()
```

## Map pre-allocation

```go
// Погано: map росте і рехешується
m := make(map[string]int)

// Добре: підказати очікуваний розмір
m := make(map[string]int, 1000)
```

## Уникати interface boxing

```go
// Кожен виклик із interface{} = потенційна алокація
var i any = 42  // int → heap

// Конкретні типи → stack (якщо не escapes)
n := 42
```

## Benchmarking tips

```go
func BenchmarkFoo(b *testing.B) {
    // Setup поза петлею:
    data := generateData(1000)

    b.ResetTimer() // скидаємо якщо setup довгий

    for b.Loop() {   // Go 1.24+: b.Loop() автоматично керує N
        process(data)
    }
}
```

```bash
go test -bench=BenchmarkFoo -benchmem -count=5 ./...
# -benchmem: показує B/op та allocs/op
# -count=5:  запустити 5 разів для стабільності
```
