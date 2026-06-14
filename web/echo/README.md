# Echo

Мінімалістичний веб-фреймворк з вбудованими middleware, bind/validate, та зручними хелперами.

## Запуск

```bash
go run main.go
```

## Ендпоінти

```
GET    /                        — health check
GET    /api/v1/users            — список користувачів
GET    /api/v1/users/:id        — користувач за ID
POST   /api/v1/users            — створити
PUT    /api/v1/users/:id        — оновити
DELETE /api/v1/users/:id        — видалити
GET    /api/v1/users/:id/posts  — пости користувача
```

## Ключові концепції

### Routing та групи

```go
e := echo.New()

// Група з префіксом і middleware
v1 := e.Group("/api/v1", middleware.Logger())
users := v1.Group("/users")
users.GET("", listUsers)
users.GET("/:id", getUser)
```

### Bind — парсинг запиту

```go
func createUser(c echo.Context) error {
    var u User
    if err := c.Bind(&u); err != nil {  // body + query + path params
        return echo.ErrBadRequest
    }
    // ...
}
```

### Відповіді

```go
c.JSON(http.StatusOK, user)           // JSON
c.JSONPretty(http.StatusOK, user, "  ") // pretty JSON
c.NoContent(http.StatusNoContent)     // 204
return echo.NewHTTPError(404, "not found") // error
```

### Кастомний middleware

```go
func myMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // before
        err := next(c)
        // after
        return err
    }
}
```

### Кастомний error handler

```go
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := http.StatusInternalServerError
    var he *echo.HTTPError
    if errors.As(err, &he) {
        code = he.Code
    }
    c.JSON(code, map[string]any{"error": err.Error()})
}
```

## Порівняння з net/http

| | net/http | Echo |
|--|----------|------|
| Залежності | 0 | ~5 |
| Path params | `r.PathValue("id")` | `c.Param("id")` |
| JSON response | вручну | `c.JSON(...)` |
| Bind body | вручну | `c.Bind(&v)` |
| Middleware | `func(Handler) Handler` | `func(HandlerFunc) HandlerFunc` |
| Groups | немає (вручну) | `e.Group("/prefix")` |
| Error handling | вручну | централізований handler |
