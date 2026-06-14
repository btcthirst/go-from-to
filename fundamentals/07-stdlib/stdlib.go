package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- strings ---

func stringsDemo() {
	fmt.Println("=== strings ===")
	s := "Hello, Gophers!"

	fmt.Println("  Contains:", strings.Contains(s, "Go"))
	fmt.Println("  HasPrefix:", strings.HasPrefix(s, "Hello"))
	fmt.Println("  ToUpper:", strings.ToUpper(s))
	fmt.Println("  TrimSpace:", strings.TrimSpace("   trim me   "))

	parts := strings.Split("a,b,c,d", ",")
	fmt.Println("  Split:", parts)
	fmt.Println("  Join:", strings.Join(parts, " | "))

	fmt.Println("  Replace:", strings.Replace("aabbcc", "b", "X", 1))
	fmt.Println("  ReplaceAll:", strings.ReplaceAll("aabbcc", "b", "X"))

	// Builder — ефективна конкатенація
	var b strings.Builder
	for i := range 5 {
		fmt.Fprintf(&b, "item%d", i)
		if i < 4 {
			b.WriteString(", ")
		}
	}
	fmt.Println("  Builder:", b.String())
}

// --- strconv ---

func strconvDemo() {
	fmt.Println("\n=== strconv ===")

	fmt.Println("  Itoa:", strconv.Itoa(42))

	if n, err := strconv.Atoi("123"); err == nil {
		fmt.Println("  Atoi:", n)
	}

	if _, err := strconv.Atoi("abc"); err != nil {
		fmt.Println("  Atoi error:", err)
	}

	if f, err := strconv.ParseFloat("3.14159", 64); err == nil {
		fmt.Printf("  ParseFloat: %.5f\n", f)
	}

	fmt.Println("  FormatFloat:", strconv.FormatFloat(3.14159, 'f', 2, 64))
	fmt.Println("  FormatBool:", strconv.FormatBool(true))
}

// --- os ---

func osDemo() {
	fmt.Println("\n=== os ===")
	fmt.Println("  os.Args:", os.Args)
	fmt.Println("  HOME:", os.Getenv("HOME"))

	// Запис і читання файлу
	tmpFile := "/tmp/go-stdlib-demo.txt"
	content := []byte("Hello from Go!\nLine 2\nLine 3\n")

	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		fmt.Println("  WriteFile error:", err)
		return
	}
	fmt.Println("  WriteFile: ok")

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		fmt.Println("  ReadFile error:", err)
		return
	}
	fmt.Printf("  ReadFile: %d bytes\n", len(data))
}

// --- bufio: читання рядок за рядком ---

func bufioDemo() {
	fmt.Println("\n=== bufio ===")
	content := "line one\nline two\nline three\n"
	r := strings.NewReader(content)
	scanner := bufio.NewScanner(r)

	for i := 1; scanner.Scan(); i++ {
		fmt.Printf("  [%d] %s\n", i, scanner.Text())
	}
}

// --- time ---

func timeDemo() {
	fmt.Println("\n=== time ===")
	now := time.Now()
	fmt.Println("  Now:", now.Format("2006-01-02 15:04:05"))
	fmt.Println("  Unix:", now.Unix())

	// Тривалості
	d := 2*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Println("  Duration:", d)
	fmt.Printf("  Hours: %.1f\n", d.Hours())

	// Таймер
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	fmt.Println("  Elapsed:", time.Since(start).Round(time.Microsecond))

	// Парсинг
	t, err := time.Parse("2006-01-02", "2025-06-14")
	if err == nil {
		fmt.Println("  Parsed:", t.Format("02 Jan 2006"))
	}
}

// --- encoding/json ---

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"in_stock"`
	Tags     []string `json:"tags,omitempty"`
	internal string  // не експортується — відсутнє в JSON
}

func jsonDemo() {
	fmt.Println("\n=== encoding/json ===")

	p := Product{
		ID:      1,
		Name:    "Gopher T-Shirt",
		Price:   29.99,
		InStock: true,
		Tags:    []string{"go", "merch"},
	}

	// Marshal
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Println("  Marshal error:", err)
		return
	}
	fmt.Println("  JSON:", string(data))

	// MarshalIndent
	pretty, _ := json.MarshalIndent(p, "  ", "  ")
	fmt.Println("  Pretty:\n ", string(pretty))

	// Unmarshal
	raw := `{"id":2,"name":"Go Book","price":49.99,"in_stock":false}`
	var p2 Product
	if err := json.Unmarshal([]byte(raw), &p2); err != nil {
		fmt.Println("  Unmarshal error:", err)
		return
	}
	fmt.Printf("  Decoded: %+v\n", p2)

	// map[string]any для динамічного JSON
	var m map[string]any
	json.Unmarshal([]byte(raw), &m)
	fmt.Printf("  price as any: %v (%T)\n", m["price"], m["price"])
}

// --- net/http: простий сервер ---

func httpServerDemo() {
	fmt.Println("\n=== net/http server (запуск на :18080) ===")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		fmt.Fprintf(w, "Hello, %s!\n", name)
	})
	mux.HandleFunc("GET /json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Запускаємо в goroutine і одразу робимо запит
	go func() {
		http.ListenAndServe(":18080", mux)
	}()
	time.Sleep(10 * time.Millisecond) // чекаємо поки сервер запуститься

	// --- HTTP client ---
	resp, err := http.Get("http://localhost:18080/hello?name=Gopher")
	if err != nil {
		fmt.Println("  GET error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  GET /hello: %s", body)

	resp2, _ := http.Get("http://localhost:18080/json")
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("  GET /json: %s", body2)
}

func main() {
	stringsDemo()
	strconvDemo()
	osDemo()
	bufioDemo()
	timeDemo()
	jsonDemo()
	httpServerDemo()
}
