package main

import (
	"fmt"
	"math"
)

// --- Structs ---

type Person struct {
	Name string
	Age  int
}

type Address struct {
	City    string
	Country string
}

// Embedding: Person в Employee
type Employee struct {
	Person
	Address
	Company string
}

// --- Methods ---

type Rectangle struct {
	Width, Height float64
}

// Value receiver
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Pointer receiver — змінює оригінал
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// --- Interfaces ---

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

func printShape(s Shape) {
	fmt.Printf("  %T: area=%.2f, perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

// --- Stringer interface (fmt.Println використовує його автоматично) ---

func (p Person) String() string {
	return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

// --- Generics (Go 1.18+) ---

func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}

func Filter[T any](slice []T, pred func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if pred(v) {
			result = append(result, v)
		}
	}
	return result
}

func main() {
	// --- Structs ---
	fmt.Println("=== Structs ===")
	p := Person{Name: "Alice", Age: 30}
	fmt.Println(p) // використовує String()

	p2 := Person{"Bob", 25} // positional (уникай — крихко)
	_ = p2

	// Анонімна структура
	point := struct{ X, Y int }{X: 3, Y: 4}
	fmt.Printf("point: %+v\n", point)

	// Pointers to structs
	pp := &Person{Name: "Carol", Age: 28}
	pp.Age++ // автоматичне dereference (*pp).Age++
	fmt.Println(pp)

	// --- Embedding ---
	fmt.Println("\n=== Embedding ===")
	e := Employee{
		Person:  Person{Name: "Dave", Age: 35},
		Address: Address{City: "Kyiv", Country: "Ukraine"},
		Company: "Acme",
	}
	fmt.Printf("Name: %s, City: %s\n", e.Name, e.City) // прямий доступ
	fmt.Println(e.Person)                                // через тип

	// --- Methods ---
	fmt.Println("\n=== Methods ===")
	r := Rectangle{Width: 5, Height: 3}
	fmt.Printf("Area=%.1f, Perimeter=%.1f\n", r.Area(), r.Perimeter())
	r.Scale(2)
	fmt.Printf("After Scale(2): %+v\n", r)

	// --- Interfaces ---
	fmt.Println("\n=== Interfaces ===")
	shapes := []Shape{
		Rectangle{3, 4},
		Circle{Radius: 5},
	}
	for _, s := range shapes {
		printShape(s)
	}

	// --- Type assertions ---
	fmt.Println("\n=== Type assertions ===")
	var s Shape = Circle{Radius: 2}
	if c, ok := s.(Circle); ok {
		fmt.Printf("Circle radius: %.1f\n", c.Radius)
	}
	if _, ok := s.(Rectangle); !ok {
		fmt.Println("Not a Rectangle")
	}

	// --- Slices ---
	fmt.Println("\n=== Slices ===")
	nums := []int{5, 3, 8, 1, 9, 2, 7}
	fmt.Printf("slice: %v, len=%d, cap=%d\n", nums, len(nums), cap(nums))

	sub := nums[2:5]
	fmt.Printf("nums[2:5] = %v\n", sub)

	// append може змінити underlying array якщо cap недостатньо
	grown := append(nums, 100)
	fmt.Printf("grown: %v\n", grown)

	// copy — справжня копія
	dst := make([]int, len(nums))
	copy(dst, nums)
	dst[0] = 999
	fmt.Printf("original: %v (незмінений)\n", nums)

	// --- Maps ---
	fmt.Println("\n=== Maps ===")
	m := map[string]int{
		"apple":  5,
		"banana": 3,
	}
	m["cherry"] = 8

	if v, ok := m["banana"]; ok {
		fmt.Printf("banana: %d\n", v)
	}
	if _, ok := m["grape"]; !ok {
		fmt.Println("grape: not found")
	}

	delete(m, "banana")
	fmt.Printf("after delete: %v\n", m)

	// --- Generics ---
	fmt.Println("\n=== Generics ===")
	doubled := Map([]int{1, 2, 3, 4, 5}, func(x int) int { return x * 2 })
	fmt.Printf("doubled: %v\n", doubled)

	evens := Filter([]int{1, 2, 3, 4, 5, 6}, func(x int) bool { return x%2 == 0 })
	fmt.Printf("evens: %v\n", evens)

	words := Map([]int{1, 2, 3}, func(x int) string { return fmt.Sprintf("item%d", x) })
	fmt.Printf("words: %v\n", words)
}
