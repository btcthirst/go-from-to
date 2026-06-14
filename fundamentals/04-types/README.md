# 04 — Типи та інтерфейси

## Structs

```go
type Person struct {
    Name string
    Age  int
}

p := Person{Name: "Alice", Age: 30}
p.Name = "Bob"

// Анонімна структура
point := struct{ X, Y int }{X: 1, Y: 2}
```

## Pointers

```go
x := 42
p := &x        // pointer to x
*p = 100       // dereference — змінює x
fmt.Println(x) // 100

// new() — виділяє пам'ять, повертає pointer
n := new(int)  // *int, *n == 0
```

> Go не має pointer arithmetic. `nil` pointer — zero value для pointers.

## Methods

```go
type Rectangle struct {
    Width, Height float64
}

// Value receiver — не змінює оригінал
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver — може змінювати оригінал
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}
```

**Коли використовувати pointer receiver:**
- Потрібно змінити значення
- Велика структура (уникнути копіювання)
- Консистентність: якщо хоч один метод — pointer, краще всі pointer

## Interfaces

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}
```

- **Implicit implementation** — не потрібно явно вказувати `implements`
- Якщо тип має всі методи інтерфейсу — він його реалізує автоматично
- Інтерфейс `interface{}` (або `any`) приймає будь-яке значення

```go
var s Shape = Rectangle{3, 4}
fmt.Println(s.Area())
```

## Type assertions та type switch

```go
var i any = "hello"

// Type assertion
s, ok := i.(string)  // ok = true, s = "hello"
n, ok := i.(int)     // ok = false, n = 0

// Type switch
switch v := i.(type) {
case string:  fmt.Println("string:", v)
case int:     fmt.Println("int:", v)
}
```

## Embedding (вбудовування)

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " says..."
}

type Dog struct {
    Animal           // вбудовування — не успадкування!
    Breed  string
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Husky"}
d.Speak()    // делегується до Animal.Speak()
d.Name       // прямий доступ до поля Animal
```

## Slice

```go
s := []int{1, 2, 3}         // літерал
s := make([]int, 5)          // len=5, cap=5
s := make([]int, 0, 10)      // len=0, cap=10

s = append(s, 4, 5)          // додати елементи
s = append(s, other...)      // злити два слайси

s[1:3]   // елементи [1,3) — нижня включена, верхня не
s[:3]    // від початку до 3
s[2:]    // від 2 до кінця
```

> Slice — це (pointer, len, cap). `copy()` — для справжнього копіювання.

## Map

```go
m := map[string]int{"a": 1, "b": 2}
m := make(map[string]int)

m["key"] = 42
v, ok := m["key"]   // ok = false якщо ключа немає
delete(m, "key")

// Ітерація (порядок не гарантований)
for k, v := range m { }
```
