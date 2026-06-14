# 02 — Управління потоком

## if / else

```go
// Звичайний
if x > 0 {
    fmt.Println("positive")
} else if x < 0 {
    fmt.Println("negative")
} else {
    fmt.Println("zero")
}

// З init-виразом (err видима тільки в блоці if/else)
if err := doSomething(); err != nil {
    log.Fatal(err)
}
```

## for — єдиний цикл у Go

```go
// C-style
for i := 0; i < 10; i++ { }

// While-style
for x < 100 { x *= 2 }

// Нескінченний цикл
for {
    if done { break }
}

// range по slice
for i, v := range slice { }
for _, v := range slice { }   // ігноруємо індекс
for i := range slice { }      // тільки індекс

// range по map
for k, v := range m { }

// range по string — ітерує по рунах (не байтах!)
for i, r := range "Привіт" { }

// range по channel — читає до закриття каналу
for v := range ch { }
```

## switch

```go
// Без умови — як if/else chain
switch {
case x < 0:  fmt.Println("negative")
case x == 0: fmt.Println("zero")
default:     fmt.Println("positive")
}

// З виразом
switch os := runtime.GOOS; os {
case "linux":   fmt.Println("Linux")
case "darwin":  fmt.Println("macOS")
default:        fmt.Printf("Other: %s\n", os)
}

// fallthrough — продовжує виконання наступного case
switch x {
case 1:
    fmt.Println("one")
    fallthrough
case 2:
    fmt.Println("one or two")
}
```

## Type switch

```go
func describe(i interface{}) {
    switch v := i.(type) {
    case int:     fmt.Printf("int: %d\n", v)
    case string:  fmt.Printf("string: %q\n", v)
    case bool:    fmt.Printf("bool: %t\n", v)
    default:      fmt.Printf("unknown: %T\n", v)
    }
}
```

## defer

Виконується перед виходом з функції. Корисний для cleanup.

```go
func readFile(path string) {
    f, err := os.Open(path)
    if err != nil { log.Fatal(err) }
    defer f.Close()   // виконається при виході з функції
    // ...
}
```

**Важливо:** декілька defer виконуються у порядку **LIFO** (стек):
```go
defer fmt.Println("first defer")   // виконається третім
defer fmt.Println("second defer")  // виконається другим
defer fmt.Println("third defer")   // виконається першим
```

**Аргументи defer обчислюються одразу** (при оголошенні), не при виконанні:
```go
x := 10
defer fmt.Println(x)  // надрукує 10, навіть якщо x зміниться
x = 20
```
