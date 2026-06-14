package main

import (
	"fmt"
	"runtime"
)

func main() {
	// --- if з init-виразом ---
	if x := compute(); x > 0 {
		fmt.Printf("compute() = %d (positive)\n", x)
	}

	// --- for: всі форми ---
	fmt.Println("\n--- for: C-style ---")
	for i := 0; i < 3; i++ {
		fmt.Printf("  i=%d\n", i)
	}

	fmt.Println("--- for: while-style ---")
	n := 1
	for n < 10 {
		n *= 2
	}
	fmt.Printf("  n=%d\n", n)

	fmt.Println("--- for: range по slice ---")
	fruits := []string{"apple", "banana", "cherry"}
	for i, f := range fruits {
		fmt.Printf("  [%d] %s\n", i, f)
	}

	fmt.Println("--- for: range по map ---")
	scores := map[string]int{"Alice": 95, "Bob": 87}
	for name, score := range scores {
		fmt.Printf("  %s: %d\n", name, score)
	}

	fmt.Println("--- for: range по string (руни, не байти) ---")
	for i, r := range "Go🚀" {
		fmt.Printf("  byte[%d] = %c (U+%04X)\n", i, r, r)
	}

	// --- switch ---
	fmt.Println("\n--- switch: з виразом ---")
	switch os := runtime.GOOS; os {
	case "linux":
		fmt.Println("  Linux")
	case "darwin":
		fmt.Println("  macOS")
	default:
		fmt.Printf("  Інша ОС: %s\n", os)
	}

	fmt.Println("--- switch: без виразу (як if-else) ---")
	hour := 14
	switch {
	case hour < 12:
		fmt.Println("  Ранок")
	case hour < 18:
		fmt.Println("  День")
	default:
		fmt.Println("  Вечір")
	}

	fmt.Println("--- switch: fallthrough ---")
	x := 1
	switch x {
	case 1:
		fmt.Println("  case 1")
		fallthrough
	case 2:
		fmt.Println("  case 1 або 2 (fallthrough)")
	case 3:
		fmt.Println("  case 3")
	}

	// --- type switch ---
	fmt.Println("\n--- type switch ---")
	values := []any{42, "hello", true, 3.14}
	for _, v := range values {
		describe(v)
	}

	// --- defer: LIFO порядок ---
	fmt.Println("\n--- defer: LIFO ---")
	deferDemo()

	// --- defer: аргументи обчислюються одразу ---
	fmt.Println("--- defer: аргументи одразу ---")
	deferArgs()
}

func compute() int { return 7 }

func describe(i any) {
	switch v := i.(type) {
	case int:
		fmt.Printf("  int: %d\n", v)
	case string:
		fmt.Printf("  string: %q\n", v)
	case bool:
		fmt.Printf("  bool: %t\n", v)
	default:
		fmt.Printf("  %T: %v\n", v, v)
	}
}

func deferDemo() {
	for i := 1; i <= 3; i++ {
		defer fmt.Printf("  defer %d\n", i)
	}
	fmt.Println("  (після реєстрації всіх defer)")
}

func deferArgs() {
	x := 10
	defer fmt.Printf("  defer бачить x=%d (значення при оголошенні)\n", x)
	x = 99
	fmt.Printf("  зараз x=%d\n", x)
}
