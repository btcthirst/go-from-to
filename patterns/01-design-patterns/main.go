package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ================================================================
// FUNCTIONAL OPTIONS
// ================================================================

type Server struct {
	host        string
	port        int
	timeout     time.Duration
	maxConns    int
	middlewares []string
}

type Option func(*Server)

func WithHost(h string) Option              { return func(s *Server) { s.host = h } }
func WithPort(p int) Option                 { return func(s *Server) { s.port = p } }
func WithTimeout(d time.Duration) Option    { return func(s *Server) { s.timeout = d } }
func WithMaxConns(n int) Option             { return func(s *Server) { s.maxConns = n } }
func WithMiddleware(name string) Option     { return func(s *Server) { s.middlewares = append(s.middlewares, name) } }

func NewServer(opts ...Option) *Server {
	s := &Server{
		host:     "localhost",
		port:     8080,
		timeout:  30 * time.Second,
		maxConns: 100,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// ================================================================
// BUILDER
// ================================================================

type QueryBuilder struct {
	table      string
	conditions []string
	orderBy    string
	limit      int
	offset     int
}

func NewQuery(table string) *QueryBuilder {
	return &QueryBuilder{table: table, limit: -1}
}

func (b *QueryBuilder) Where(condition string) *QueryBuilder {
	b.conditions = append(b.conditions, condition)
	return b
}

func (b *QueryBuilder) OrderBy(field string) *QueryBuilder {
	b.orderBy = field
	return b
}

func (b *QueryBuilder) Limit(n int) *QueryBuilder {
	b.limit = n
	return b
}

func (b *QueryBuilder) Offset(n int) *QueryBuilder {
	b.offset = n
	return b
}

func (b *QueryBuilder) Build() (string, error) {
	if b.table == "" {
		return "", errors.New("table is required")
	}
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(b.table)
	if len(b.conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.conditions, " AND "))
	}
	if b.orderBy != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(b.orderBy)
	}
	if b.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	if b.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}
	return sb.String(), nil
}

// ================================================================
// MIDDLEWARE CHAIN
// ================================================================

type Request struct {
	Method string
	Path   string
}

type Response struct {
	Status int
	Body   string
}

type Handler func(ctx context.Context, req Request) (Response, error)
type Middleware func(Handler) Handler

func Chain(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func Logging(next Handler) Handler {
	return func(ctx context.Context, req Request) (Response, error) {
		fmt.Printf("  [LOG] → %s %s\n", req.Method, req.Path)
		resp, err := next(ctx, req)
		fmt.Printf("  [LOG] ← %d\n", resp.Status)
		return resp, err
	}
}

func Recover(next Handler) Handler {
	return func(ctx context.Context, req Request) (resp Response, err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  [RECOVER] panic: %v\n", r)
				resp = Response{Status: 500, Body: "internal error"}
				err = nil
			}
		}()
		return next(ctx, req)
	}
}

func Auth(token string) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req Request) (Response, error) {
			if token == "" {
				return Response{Status: 401, Body: "unauthorized"}, nil
			}
			ctx = context.WithValue(ctx, "user", "authenticated")
			return next(ctx, req)
		}
	}
}

// ================================================================
// REPOSITORY
// ================================================================

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*User, error)
	Save(ctx context.Context, u *User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*User, error)
}

// InMemory — для тестів і прикладів
type InMemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{users: make(map[int]*User), nextID: 1}
}

func (r *InMemoryUserRepo) FindByID(_ context.Context, id int) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d: not found", id)
	}
	return u, nil
}

func (r *InMemoryUserRepo) Save(_ context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == 0 {
		u.ID = r.nextID
		r.nextID++
	}
	r.users[u.ID] = u
	return nil
}

func (r *InMemoryUserRepo) Delete(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("user %d: not found", id)
	}
	delete(r.users, id)
	return nil
}

func (r *InMemoryUserRepo) List(_ context.Context) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

// Decorator: кешований репозиторій поверх будь-якого UserRepository
type CachedUserRepo struct {
	inner UserRepository
	cache map[int]*User
	mu    sync.RWMutex
}

func WithCache(r UserRepository) *CachedUserRepo {
	return &CachedUserRepo{inner: r, cache: make(map[int]*User)}
}

func (r *CachedUserRepo) FindByID(ctx context.Context, id int) (*User, error) {
	r.mu.RLock()
	if u, ok := r.cache[id]; ok {
		r.mu.RUnlock()
		fmt.Printf("  [CACHE] hit id=%d\n", id)
		return u, nil
	}
	r.mu.RUnlock()

	u, err := r.inner.FindByID(ctx, id)
	if err == nil {
		r.mu.Lock()
		r.cache[id] = u
		r.mu.Unlock()
		fmt.Printf("  [CACHE] miss id=%d, stored\n", id)
	}
	return u, err
}

func (r *CachedUserRepo) Save(ctx context.Context, u *User) error {
	err := r.inner.Save(ctx, u)
	if err == nil {
		r.mu.Lock()
		delete(r.cache, u.ID) // інвалідація
		r.mu.Unlock()
	}
	return err
}

func (r *CachedUserRepo) Delete(ctx context.Context, id int) error {
	err := r.inner.Delete(ctx, id)
	if err == nil {
		r.mu.Lock()
		delete(r.cache, id)
		r.mu.Unlock()
	}
	return err
}

func (r *CachedUserRepo) List(ctx context.Context) ([]*User, error) {
	return r.inner.List(ctx)
}

// ================================================================
// SINGLETON (sync.Once)
// ================================================================

type Config struct {
	DSN     string
	MaxConn int
}

var (
	cfgOnce sync.Once
	cfg     *Config
)

func GetConfig() *Config {
	cfgOnce.Do(func() {
		fmt.Println("  [SINGLETON] ініціалізація Config...")
		cfg = &Config{DSN: "postgres://localhost/mydb", MaxConn: 10}
	})
	return cfg
}

// ================================================================
// OBSERVER (через channels)
// ================================================================

type Event struct {
	Topic   string
	Payload any
}

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[string][]chan Event)}
}

func (b *EventBus) Subscribe(topic string) <-chan Event {
	ch := make(chan Event, 10)
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Publish(topic string, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e := Event{Topic: topic, Payload: payload}
	for _, ch := range b.subscribers[topic] {
		select {
		case ch <- e:
		default: // не блокуємо якщо підписник не встигає
		}
	}
}

func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, subs := range b.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
}

// ================================================================
// MAIN
// ================================================================

func main() {
	// Functional Options
	fmt.Println("=== Functional Options ===")
	s1 := NewServer()
	s2 := NewServer(WithPort(9090), WithTimeout(5*time.Second), WithMiddleware("logger"), WithMiddleware("auth"))
	fmt.Printf("  default: host=%s port=%d timeout=%s\n", s1.host, s1.port, s1.timeout)
	fmt.Printf("  custom:  host=%s port=%d timeout=%s mw=%v\n", s2.host, s2.port, s2.timeout, s2.middlewares)

	// Builder
	fmt.Println("\n=== Builder ===")
	q, err := NewQuery("users").
		Where("age > 18").
		Where("active = true").
		OrderBy("name").
		Limit(10).
		Offset(20).
		Build()
	fmt.Printf("  query: %s (err=%v)\n", q, err)

	_, err = (&QueryBuilder{}).Build()
	fmt.Printf("  empty builder: %v\n", err)

	// Middleware
	fmt.Println("\n=== Middleware Chain ===")
	base := Handler(func(_ context.Context, req Request) (Response, error) {
		return Response{Status: 200, Body: "ok"}, nil
	})
	chained := Chain(base, Logging, Recover, Auth("my-token"))
	resp, _ := chained(context.Background(), Request{Method: "GET", Path: "/api/users"})
	fmt.Printf("  response: %+v\n", resp)

	// Repository + Decorator
	fmt.Println("\n=== Repository + Cache Decorator ===")
	ctx := context.Background()
	repo := WithCache(NewInMemoryUserRepo())
	repo.Save(ctx, &User{Name: "Alice", Email: "alice@example.com"})
	repo.Save(ctx, &User{Name: "Bob", Email: "bob@example.com"})

	u, _ := repo.FindByID(ctx, 1) // miss → store
	u, _ = repo.FindByID(ctx, 1) // hit
	fmt.Printf("  user: %+v\n", u)

	// Singleton
	fmt.Println("\n=== Singleton ===")
	c1 := GetConfig()
	c2 := GetConfig() // ініціалізація не повторюється
	c3 := GetConfig()
	fmt.Printf("  same instance: %t\n", c1 == c2 && c2 == c3)
	fmt.Printf("  config: %+v\n", *c1)

	// Observer
	fmt.Println("\n=== Observer (EventBus) ===")
	bus := NewEventBus()
	userEvents := bus.Subscribe("user.created")
	emailEvents := bus.Subscribe("user.created")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e := <-userEvents
		fmt.Printf("  [user-service]  отримав: %v\n", e.Payload)
	}()
	go func() {
		defer wg.Done()
		e := <-emailEvents
		fmt.Printf("  [email-service] отримав: %v\n", e.Payload)
	}()

	bus.Publish("user.created", map[string]string{"name": "Charlie", "email": "charlie@example.com"})
	wg.Wait()
}
