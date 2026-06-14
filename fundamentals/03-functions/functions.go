package main

import (
	"errors"
	"fmt"
)

// --- Базова функція ---

func add(a, b int) int {
	return a + b
}

// Кілька аргументів одного типу — тип вказується раз в кінці
func addThree(a, b, c int) int {
	return a + b + c
}

// --- Кілька значень повернення ---

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// --- Іменовані значення повернення ---

func minMax(nums []int) (min, max int) {
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return // naked return
}

// --- Variadic ---

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// --- Функції як значення ---

func apply(nums []int, f func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = f(n)
	}
	return result
}

// --- Closure: лічильник ---

func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// --- Closure: adder factory ---

func makeAdder(x int) func(int) int {
	return func(y int) int {
		return x + y
	}
}

// --- Рекурсія ---

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// --- Рекурсія: ітеративна версія (краще для великих n) ---

func fibIterative(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for range n - 1 {
		a, b = b, a+b
	}
	return b
}

// --- init ---

func init() {
	// Виконується автоматично перед main()
	// Тут можна ініціалізувати глобальні змінні, реєструвати щось тощо
}

func main() {
	// Базові виклики
	fmt.Printf("add(3,4) = %d\n", add(3, 4))
	fmt.Printf("addThree(1,2,3) = %d\n", addThree(1, 2, 3))

	// Кілька значень повернення
	if result, err := divide(10, 3); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("10/3 = %.4f\n", result)
	}
	if _, err := divide(5, 0); err != nil {
		fmt.Println("divide by zero:", err)
	}

	// Іменовані значення
	nums := []int{5, 2, 8, 1, 9, 3}
	min, max := minMax(nums)
	fmt.Printf("min=%d, max=%d\n", min, max)

	// Variadic
	fmt.Printf("sum(1,2,3) = %d\n", sum(1, 2, 3))
	slice := []int{10, 20, 30}
	fmt.Printf("sum(slice...) = %d\n", sum(slice...))

	// Функції як значення
	doubled := apply([]int{1, 2, 3}, func(x int) int { return x * 2 })
	fmt.Printf("doubled: %v\n", doubled)

	// Closures — незалежні стани
	c1 := makeCounter()
	c2 := makeCounter()
	fmt.Printf("c1: %d, %d, %d\n", c1(), c1(), c1())
	fmt.Printf("c2: %d (незалежний)\n", c2())

	add5 := makeAdder(5)
	add10 := makeAdder(10)
	fmt.Printf("add5(3)=%d, add10(3)=%d\n", add5(3), add10(3))

	// Рекурсія
	fmt.Printf("fibonacci(10) = %d\n", fibonacci(10))
	fmt.Printf("fibIterative(10) = %d\n", fibIterative(10))

	// Анонімна функція (IIFE)
	result := func(a, b int) int { return a * b }(6, 7)
	fmt.Printf("6*7 = %d\n", result)
}
