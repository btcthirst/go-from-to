package main

import "fmt"

// --- Константи з iota ---

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

const (
	KB = 1 << (10 * (iota + 1))
	MB
	GB
)

// --- Змінні на рівні пакету ---

var packageLevel = "доступна в усьому пакеті"

func main() {
	// --- Оголошення змінних ---
	var x int
	var s = "hello"
	n := 42
	fmt.Printf("x=%d (zero value), s=%q, n=%d\n", x, s, n)

	// Кілька змінних одночасно
	a, b := 10, 20
	fmt.Printf("a=%d, b=%d\n", a, b)

	// Swap без temp
	a, b = b, a
	fmt.Printf("після swap: a=%d, b=%d\n", a, b)

	// --- Zero values ---
	var i int
	var f float64
	var bl bool
	var str string
	fmt.Printf("zero values: int=%d, float=%f, bool=%t, string=%q\n", i, f, bl, str)

	// --- Типи ---
	var myByte byte = 'A'          // uint8
	var myRune rune = '€'          // int32, Unicode code point
	fmt.Printf("byte=%d char=%c, rune=%d char=%c\n", myByte, myByte, myRune, myRune)

	// --- Конвертація типів (явна) ---
	var pi float64 = 3.14159
	var piInt int = int(pi) // втрачаємо дробову частину
	fmt.Printf("float %.5f -> int %d\n", pi, piInt)

	// --- iota ---
	fmt.Printf("Sunday=%d, Monday=%d, Saturday=%d\n", Sunday, Monday, Saturday)
	fmt.Printf("KB=%d, MB=%d, GB=%d\n", KB, MB, GB)

	// --- Printf формати ---
	type Point struct{ X, Y int }
	p := Point{1, 2}
	fmt.Printf("%%v  = %v\n", p)
	fmt.Printf("%%+v = %+v\n", p)
	fmt.Printf("%%#v = %#v\n", p)
	fmt.Printf("%%T  = %T\n", p)
	fmt.Printf("hex = %x, bin = %b\n", 255, 255)

	// --- Рядки та руни ---
	word := "Привіт"
	fmt.Printf("len(%q) = %d bytes, %d runes\n", word, len(word), len([]rune(word)))

	// Ітерація по рунах (не байтах)
	for i, r := range word {
		fmt.Printf("  [%d] %c (%d)\n", i, r, r)
	}
}
