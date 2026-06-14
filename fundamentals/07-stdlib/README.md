# 07 — Стандартна бібліотека

## fmt

```go
fmt.Println("hello", 42)           // рядок + пробіл + \n
fmt.Printf("%.2f\n", 3.14159)      // форматований вивід
fmt.Sprintf("user: %s", name)      // повертає рядок
fmt.Errorf("not found: %w", err)   // повертає error з wrapping

fmt.Scan(&x)                        // читати з stdin
fmt.Scanf("%d %s", &n, &s)
fmt.Sscanf("42 hello", "%d %s", &n, &s)
```

## strings

```go
strings.Contains("hello", "ell")     // true
strings.HasPrefix("hello", "hel")    // true
strings.HasSuffix("hello", "llo")    // true
strings.Index("hello", "ll")         // 2
strings.Count("hello", "l")          // 2

strings.ToUpper("hello")             // "HELLO"
strings.ToLower("HELLO")             // "hello"
strings.TrimSpace("  hi  ")          // "hi"
strings.Trim("--hi--", "-")          // "hi"

strings.Split("a,b,c", ",")          // ["a","b","c"]
strings.Join([]string{"a","b"}, "-") // "a-b"
strings.Replace("aaa", "a", "b", 2)  // "bba"
strings.ReplaceAll("aaa", "a", "b")  // "bbb"

// Builder — ефективна конкатенація
var b strings.Builder
for i := range 5 {
    fmt.Fprintf(&b, "item%d ", i)
}
s := b.String()
```

## strconv

```go
strconv.Itoa(42)             // "42"
strconv.Atoi("42")           // 42, nil
strconv.ParseFloat("3.14", 64) // 3.14, nil
strconv.ParseBool("true")    // true, nil
strconv.FormatFloat(3.14, 'f', 2, 64) // "3.14"
```

## os

```go
os.Args                           // аргументи командного рядка
os.Getenv("PATH")                 // змінна середовища
os.Setenv("KEY", "value")
os.Exit(1)                        // вийти з кодом

// Файли
f, err := os.Open("file.txt")     // тільки читання
f, err := os.Create("file.txt")   // записування (очищує)
f, err := os.OpenFile("f.txt", os.O_APPEND|os.O_WRONLY, 0644)
defer f.Close()

data, err := os.ReadFile("file.txt")      // весь файл у []byte
err := os.WriteFile("out.txt", data, 0644)
```

## io / bufio

```go
// io.Reader / io.Writer — основні інтерфейси
io.Copy(dst, src)      // копіює з Reader в Writer
io.ReadAll(r)          // читає все до EOF

// bufio — буферизований ввід/вивід
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    line := scanner.Text()
}

writer := bufio.NewWriter(f)
writer.WriteString("hello\n")
writer.Flush()
```

## time

```go
time.Now()                              // поточний час
time.Now().Format("2006-01-02 15:04:05") // форматування (магічна дата!)
time.Parse("2006-01-02", "2025-01-15") // парсинг

time.Since(start)          // тривалість від start
time.Until(deadline)       // тривалість до deadline

d := 5 * time.Second
time.Sleep(d)

timer := time.NewTimer(d)
<-timer.C                  // чекає d і отримує час

ticker := time.NewTicker(1 * time.Second)
for t := range ticker.C {
    fmt.Println(t)
}
ticker.Stop()
```

> **Важливо:** магічна дата Go — `Mon Jan 2 15:04:05 MST 2006` (або `2006-01-02 15:04:05`). Саме ці числа використовуються для форматування.

## encoding/json

```go
// Marshal
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age,omitempty"`
    pass string  // не експортується — не включається
}

data, err := json.Marshal(user)
pretty, err := json.MarshalIndent(user, "", "  ")

// Unmarshal
var u User
err := json.Unmarshal(data, &u)

// Encoder/Decoder (для потоків)
json.NewEncoder(w).Encode(user)
json.NewDecoder(r).Decode(&user)
```

## net/http

```go
// HTTP server
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello!")
})
http.ListenAndServe(":8080", nil)

// HTTP client
resp, err := http.Get("https://example.com")
defer resp.Body.Close()
body, err := io.ReadAll(resp.Body)

// POST з JSON
data, _ := json.Marshal(payload)
resp, err := http.Post(url, "application/json", bytes.NewReader(data))
```
