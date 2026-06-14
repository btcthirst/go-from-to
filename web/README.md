# Веб-розробка на Go

Всі три приклади реалізують **однаковий** CRUD API (`/api/.../users`) — щоб легко порівнювати підходи.

| | Опис |
|--|------|
| [net/http](./net-http/) | Голий Go, без залежностей. Go 1.22+ routing |
| [Echo](./echo/) | Мінімалістичний фреймворк, вбудовані middleware та bind |
| [Chi](./chi/) | Легкий роутер, 100% сумісний зі стандартним `net/http` |

## Порівняння

| | net/http | Echo | Chi |
|--|----------|------|-----|
| Залежності | 0 | ~5 | 0 |
| Path params | `r.PathValue("id")` | `c.Param("id")` | `chi.URLParam(r, "id")` |
| JSON response | вручну | `c.JSON(...)` | вручну або render |
| Bind body | вручну | `c.Bind(&v)` | вручну |
| Middleware | `func(Handler) Handler` | `func(HandlerFunc) HandlerFunc` | `func(Handler) Handler` |
| Route groups | вручну | `e.Group("/v1")` | `r.Route("/v1", ...)` |
| Сумісність зі stdlib | повна | часткова | повна |
| Коли обирати | бібліотеки, мінімум deps | швидкий старт | stdlib-сумісність + router |

## Запуск

```bash
go run web/net-http/main.go   # :8080
go run web/echo/main.go       # :8080
go run web/chi/main.go        # :8080
```
