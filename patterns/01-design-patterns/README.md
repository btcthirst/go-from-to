# 01 — Design Patterns у Go

Go не має класичного ООП, тому класичні патерни реалізуються по-своєму.

## Functional Options

Найпоширеніший Go-патерн для конфігурації зі значеннями за замовчуванням:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithPort(p int) Option   { return func(s *Server) { s.port = p } }
func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080, timeout: 30 * time.Second}
    for _, o := range opts {
        o(s)
    }
    return s
}

// Використання
srv := NewServer(WithPort(9090), WithTimeout(5*time.Second))
```

**Переваги:** backward-compatible (нові опції не ламають API), самодокументовані.

## Builder

Коли об'єкт будується поетапно і потребує валідації:

```go
type QueryBuilder struct { table, where string; limit int }

func (b *QueryBuilder) From(t string) *QueryBuilder  { b.table = t; return b }
func (b *QueryBuilder) Where(w string) *QueryBuilder { b.where = w; return b }
func (b *QueryBuilder) Limit(n int) *QueryBuilder    { b.limit = n; return b }
func (b *QueryBuilder) Build() (string, error)       { /* validate + build */ }

q, err := (&QueryBuilder{}).From("users").Where("age > 18").Limit(10).Build()
```

## Middleware Chain

```go
type Handler func(ctx context.Context, req Request) (Response, error)
type Middleware func(Handler) Handler

func Chain(h Handler, mws ...Middleware) Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

// Кожен middleware: before → next → after
func Logging(next Handler) Handler {
    return func(ctx context.Context, req Request) (Response, error) {
        log.Printf("→ %s", req.Method)
        resp, err := next(ctx, req)
        log.Printf("← %d", resp.Status)
        return resp, err
    }
}
```

## Repository (Data Access Abstraction)

```go
type UserRepository interface {
    FindByID(ctx context.Context, id int) (*User, error)
    Save(ctx context.Context, u *User) error
    Delete(ctx context.Context, id int) error
}

// Production: SQL implementation
type PostgresUserRepo struct { db *sql.DB }
func (r *PostgresUserRepo) FindByID(...) (*User, error) { /* query */ }

// Test: in-memory implementation
type InMemoryUserRepo struct { users map[int]*User }
func (r *InMemoryUserRepo) FindByID(...) (*User, error) { /* map lookup */ }
```

## Singleton (sync.Once)

```go
var (
    once     sync.Once
    instance *DB
)

func GetDB() *DB {
    once.Do(func() { instance = connect() })
    return instance
}
```

## Observer (через channels)

```go
type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}

func (b *EventBus) Subscribe(topic string) <-chan Event {
    ch := make(chan Event, 10)
    b.mu.Lock()
    b.subscribers[topic] = append(b.subscribers[topic], ch)
    b.mu.Unlock()
    return ch
}

func (b *EventBus) Publish(topic string, e Event) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, ch := range b.subscribers[topic] {
        ch <- e
    }
}
```

## Decorator (через interface wrapping)

```go
// Додаємо кешування без зміни оригіналу
type CachedUserRepo struct {
    inner UserRepository
    cache map[int]*User
    mu    sync.RWMutex
}

func (r *CachedUserRepo) FindByID(ctx context.Context, id int) (*User, error) {
    r.mu.RLock()
    if u, ok := r.cache[id]; ok { r.mu.RUnlock(); return u, nil }
    r.mu.RUnlock()

    u, err := r.inner.FindByID(ctx, id)
    if err == nil {
        r.mu.Lock()
        r.cache[id] = u
        r.mu.Unlock()
    }
    return u, err
}
```
