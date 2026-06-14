# 06 — Помилки

## Інтерфейс error

```go
type error interface {
    Error() string
}
```

Будь-який тип з методом `Error() string` реалізує `error`.

## Базовий патерн

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}
```

> **Завжди** перевіряй `err != nil`. Ігнорування через `_` — тільки коли ти свідомо відкидаєш помилку.

## Створення помилок

```go
// Проста помилка
err := errors.New("something went wrong")

// З форматуванням
err := fmt.Errorf("user %d not found", userID)

// З wrapping (%w — дозволяє errors.Is/As)
err := fmt.Errorf("db query: %w", originalErr)
```

## Кастомні типи помилок

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s — %s", e.Field, e.Message)
}
```

## errors.Is і errors.As

```go
var ErrNotFound = errors.New("not found")

// errors.Is — перевіряє sentinel помилку (через ланцюг wrapping)
if errors.Is(err, ErrNotFound) {
    // обробити "not found"
}

// errors.As — отримати конкретний тип (через ланцюг wrapping)
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println("invalid field:", ve.Field)
}
```

> `errors.Is` і `errors.As` розгортають ланцюги `%w`, тому вони кращі за `err == ErrNotFound` чи type assertion напряму.

## Sentinel errors

```go
var (
    ErrNotFound   = errors.New("not found")
    ErrForbidden  = errors.New("forbidden")
    ErrTimeout    = errors.New("timeout")
)
```

## panic і recover

Для **дійсно** неочікуваних ситуацій (не замість error):

```go
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("expected positive, got %d", n))
    }
    return n
}

// recover — тільки в defer
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    return a / b, nil
}
```

> Використовуй `panic` тільки для програмних помилок (неможливих станів). Для бізнес-логіки — завжди `error`.

## Wrap / Unwrap

```go
// Wrapping
err1 := errors.New("root cause")
err2 := fmt.Errorf("layer 2: %w", err1)
err3 := fmt.Errorf("layer 3: %w", err2)

// Unwrapping
errors.Is(err3, err1)  // true — розгортає через весь ланцюг
errors.Unwrap(err3)    // err2

// Кілька обгорток (Go 1.20+)
err := fmt.Errorf("combined: %w and %w", err1, err2)
errors.Is(err, err1) // true
errors.Is(err, err2) // true
```
