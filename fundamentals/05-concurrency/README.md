# 05 — Конкурентність

> "Don't communicate by sharing memory; share memory by communicating." — Go Proverb

## Goroutines

Легковагові потоки, що управляються рантаймом Go (не OS threads):

```go
go func() {
    fmt.Println("runs concurrently")
}()

go doWork(arg)  // будь-яку функцію можна запустити як goroutine
```

> Програма завершується коли завершується `main()` — goroutines, що ще виконуються, вбиваються.

## Channels

Основний механізм комунікації між goroutines:

```go
ch := make(chan int)        // небуферизований
ch := make(chan int, 10)    // буферизований (не блокує до заповнення)

ch <- 42      // відправити (блокує якщо небуферизований і ніхто не читає)
v := <-ch     // отримати (блокує якщо порожній)
v, ok := <-ch // ok = false якщо канал закритий і порожній

close(ch)     // закрити канал (тільки відправник!)

// range читає до закриття каналу
for v := range ch {
    fmt.Println(v)
}
```

**Правила:**
- Відправляти в закритий канал → panic
- Отримувати з закритого → отримуєш zero value (ok = false)
- Закривати вже закритий → panic

## select

Чекає на кількох каналах одночасно:

```go
select {
case v := <-ch1:
    fmt.Println("from ch1:", v)
case v := <-ch2:
    fmt.Println("from ch2:", v)
case ch3 <- data:
    fmt.Println("sent to ch3")
default:
    fmt.Println("no channel ready (non-blocking)")
}
```

## sync.WaitGroup

Чекати на завершення групи goroutines:

```go
var wg sync.WaitGroup

for i := range 5 {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}

wg.Wait() // блокує до wg counter == 0
```

## sync.Mutex

Захист від concurrent доступу до спільних даних:

```go
type SafeCounter struct {
    mu sync.Mutex
    v  map[string]int
}

func (c *SafeCounter) Inc(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.v[key]++
}
```

> `sync.RWMutex` — коли читань багато, записів мало: `RLock()`/`RUnlock()` для читання.

## sync.Once

Виконати щось рівно один раз:

```go
var once sync.Once
var instance *DB

func getInstance() *DB {
    once.Do(func() {
        instance = &DB{...}
    })
    return instance
}
```

## context

Скасування, таймаути, передача значень між goroutines:

```go
// Таймаут
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Скасування
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(2 * time.Second)
    cancel() // сигнал скасування
}()

// Перевірка в goroutine
select {
case <-ctx.Done():
    fmt.Println("cancelled:", ctx.Err())
    return
case result := <-work:
    fmt.Println(result)
}
```

## Патерни

### Fan-out / Fan-in
```go
// Fan-out: один вхідний канал → багато workers
// Fan-in: багато каналів → один вихідний канал
```

### Worker Pool
```go
jobs := make(chan int, 100)
results := make(chan int, 100)

for w := range 3 { // 3 workers
    go worker(w, jobs, results)
}
```

### Pipeline
```go
func generate(nums ...int) <-chan int { ... }
func square(in <-chan int) <-chan int { ... }

c := generate(2, 3)
out := square(c)
```
