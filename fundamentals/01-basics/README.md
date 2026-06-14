# 01 — Базові концепції

## Змінні

```go
var x int         // оголошення з типом (zero value = 0)
var s = "hello"   // оголошення з ініціалізацією (тип виводиться)
n := 42           // коротке оголошення (тільки всередині функцій)

var a, b, c int            // кілька змінних одного типу
x, y := 10, "world"        // кілька змінних різних типів
```

> `:=` не можна використовувати на рівні пакету — тільки всередині функцій.

## Zero Values

Кожен тип має значення за замовчуванням:

| Тип | Zero value |
|-----|-----------|
| `int`, `float64` | `0` |
| `bool` | `false` |
| `string` | `""` |
| `pointer`, `slice`, `map`, `channel`, `func` | `nil` |

## Константи

```go
const Pi = 3.14159
const (
    StatusOK    = 200
    StatusNotFound = 404
)
```

### iota — лічильник констант

```go
const (
    Sunday = iota   // 0
    Monday          // 1
    Tuesday         // 2
)

const (
    KB = 1 << (10 * (iota + 1))  // 1024
    MB                             // 1048576
    GB                             // 1073741824
)
```

## Базові типи

```
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte   // аліас для uint8
rune   // аліас для int32 (Unicode code point)

float32 float64
complex64 complex128
```

> `int` — розмір залежить від платформи (32 або 64 біти). Для індексів і розмірів завжди використовуй `int`.

## Конвертація типів

Go **не робить неявних конвертацій**. Потрібно явно:

```go
var i int = 42
var f float64 = float64(i)
var u uint = uint(f)

s := string(rune(65))  // "A"  (через rune, не int)
```

## Printf — формати виводу

```
%v   — значення у форматі за замовчуванням
%+v  — структура з іменами полів
%#v  — Go-синтаксис значення
%T   — тип значення
%d   — ціле число (decimal)
%f   — float (%.2f — 2 знаки після коми)
%s   — рядок
%q   — рядок у лапках
%p   — pointer (адреса)
%b   — binary
%x   — hex
```
