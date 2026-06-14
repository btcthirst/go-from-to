# net/http — голий Go

Стандартна бібліотека без залежностей. Go 1.22+ має покращений `ServeMux` з методами та параметрами шляху.

## Запуск

```bash
go run main.go
```

## Ендпоінти

```
GET  /                        — головна
GET  /api/users               — список користувачів
GET  /api/users/{id}          — користувач за ID
POST /api/users               — створити користувача
PUT  /api/users/{id}          — оновити
DELETE /api/users/{id}        — видалити
GET  /api/users/{id}/posts    — вкладені ресурси
```

## Ключові концепції

### Routing (Go 1.22+)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", handler)  // метод + шлях + параметр
r.PathValue("id")                                // отримати параметр
```

### Middleware

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

### JSON відповіді

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}
```

## Коли обирати net/http

- Немає зовнішніх залежностей
- Максимальна сумісність і стабільність
- Бібліотеки (не хочеш нав'язувати фреймворк залежностям)
- Коли достатньо простого routing
