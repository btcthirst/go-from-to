# 03 — Функції

## Базовий синтаксис

```go
func add(a, b int) int {
    return a + b
}

// Кілька аргументів одного типу — тип вказується раз
func minMax(a, b, c int) (int, int) {
    // ...
}
```

## Кілька значень, що повертаються

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 3)
if err != nil {
    log.Fatal(err)
}
```

## Іменовані значення повернення

```go
func stats(nums []int) (min, max, sum int) {
    min, max = nums[0], nums[0]
    for _, n := range nums {
        if n < min { min = n }
        if n > max { max = n }
        sum += n
    }
    return // "naked return" — повертає min, max, sum
}
```

> Naked return знижує читабельність у довгих функціях. Використовуй обережно.

## Variadic функції

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum(1, 2, 3)          // 6
sum([]int{1,2,3}...)  // розпаковка слайсу
```

## Функції як значення

```go
// Тип функції
var fn func(int) int

// Присвоєння
fn = func(x int) int { return x * 2 }

// Передача як аргумент
func apply(nums []int, f func(int) int) []int {
    result := make([]int, len(nums))
    for i, n := range nums {
        result[i] = f(n)
    }
    return result
}
```

## Closures (замикання)

Функція, що «захоплює» змінні зовнішнього середовища:

```go
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

c := counter()
c() // 1
c() // 2
c() // 3

c2 := counter() // незалежний лічильник
c2() // 1
```

## defer + closures

```go
func measure(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", name, time.Since(start))
    }
}

func process() {
    defer measure("process")()
    // ...
}
```

## Рекурсія

```go
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
```

> Go підтримує хвостову рекурсію синтаксично, але компілятор її не оптимізує. Для глибокої рекурсії — використовуй ітерацію.

## init()

Спеціальна функція, що виконується автоматично при ініціалізації пакету:

```go
func init() {
    // реєстрація, налаштування, валідація
}
```
- Пакет може мати кілька `init()`
- Виконуються в порядку оголошення
- Не можна викликати вручну
