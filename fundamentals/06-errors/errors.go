package main

import (
	"errors"
	"fmt"
)

// --- Sentinel errors ---

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

// --- Кастомний тип помилки ---

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: field %q — %s", e.Field, e.Message)
}

// --- Функція з wrapping ---

type User struct {
	ID   int
	Name string
	Age  int
}

func findUser(id int) (*User, error) {
	users := map[int]*User{
		1: {ID: 1, Name: "Alice", Age: 30},
		2: {ID: 2, Name: "Bob", Age: 17},
	}
	u, ok := users[id]
	if !ok {
		return nil, fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
	}
	return u, nil
}

func validateUser(u *User) error {
	if u.Age < 18 {
		return &ValidationError{Field: "age", Message: "must be 18+"}
	}
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "required"}
	}
	return nil
}

func processUser(id int) error {
	u, err := findUser(id)
	if err != nil {
		return fmt.Errorf("processUser: %w", err)
	}
	if err := validateUser(u); err != nil {
		return fmt.Errorf("processUser: %w", err)
	}
	fmt.Printf("  ✓ user %q processed\n", u.Name)
	return nil
}

// --- panic / recover ---

func mustPositive(n int) int {
	if n <= 0 {
		panic(fmt.Sprintf("expected positive, got %d", n))
	}
	return n
}

func safeDiv(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("safeDiv recovered: %v", r)
		}
	}()
	return a / b, nil
}

// --- Errors.Join (Go 1.20+) ---

func validateAll(u *User) error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, &ValidationError{Field: "name", Message: "required"})
	}
	if u.Age < 18 {
		errs = append(errs, &ValidationError{Field: "age", Message: "must be 18+"})
	}
	return errors.Join(errs...)
}

func main() {
	// --- errors.Is ---
	fmt.Println("=== errors.Is ===")
	err := processUser(99)
	fmt.Println("  error:", err)
	fmt.Println("  is ErrNotFound:", errors.Is(err, ErrNotFound))
	fmt.Println("  is ErrForbidden:", errors.Is(err, ErrForbidden))

	// --- errors.As ---
	fmt.Println("\n=== errors.As ===")
	err = processUser(2) // user з age < 18
	fmt.Println("  error:", err)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("  ValidationError: field=%q msg=%q\n", ve.Field, ve.Message)
	}

	// --- Успішний case ---
	fmt.Println("\n=== Success ===")
	if err := processUser(1); err != nil {
		fmt.Println("  error:", err)
	}

	// --- panic / recover ---
	fmt.Println("\n=== panic/recover ===")
	fmt.Println("  mustPositive(5):", mustPositive(5))

	result, err := safeDiv(10, 2)
	fmt.Printf("  10/2 = %d, err = %v\n", result, err)

	result, err = safeDiv(10, 0)
	fmt.Printf("  10/0: result=%d, err=%v\n", result, err)

	// Перехоплення panic через recover
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("  recovered from panic:", r)
			}
		}()
		mustPositive(-1)
	}()

	// --- errors.Join ---
	fmt.Println("\n=== errors.Join ===")
	bad := &User{Name: "", Age: 10}
	if err := validateAll(bad); err != nil {
		fmt.Println("  combined error:", err)
		// errors.Is/As працює через joined errors
		var ve *ValidationError
		fmt.Println("  has ValidationError:", errors.As(err, &ve))
	}

	// --- Wrapping ланцюг ---
	fmt.Println("\n=== Unwrap chain ===")
	root := errors.New("root cause")
	wrapped := fmt.Errorf("middle: %w", root)
	top := fmt.Errorf("top: %w", wrapped)

	fmt.Println("  top:", top)
	fmt.Println("  Is root:", errors.Is(top, root))
	fmt.Println("  Unwrap:", errors.Unwrap(top))
	fmt.Println("  Unwrap x2:", errors.Unwrap(errors.Unwrap(top)))
}
