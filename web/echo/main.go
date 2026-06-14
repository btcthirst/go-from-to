package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// --- Models ---

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- In-memory store (ідентичний net-http прикладу) ---

type Store struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
}

func NewStore() *Store {
	s := &Store{users: make(map[int]User), nextID: 1}
	s.users[1] = User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s.users[2] = User{ID: 2, Name: "Bob", Email: "bob@example.com"}
	s.nextID = 3
	return s
}

func (s *Store) GetAll() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *Store) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *Store) Create(u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	u.ID = s.nextID
	s.nextID++
	s.users[u.ID] = u
	return u
}

func (s *Store) Update(id int, u User) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return User{}, false
	}
	u.ID = id
	s.users[id] = u
	return u, true
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return false
	}
	delete(s.users, id)
	return true
}

// --- Handlers ---

type Handler struct {
	store *Store
}

func (h *Handler) listUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, h.store.GetAll())
}

func (h *Handler) getUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	user, ok := h.store.Get(id)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user %d not found", id))
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) createUser(c echo.Context) error {
	var u User
	if err := c.Bind(&u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if u.Name == "" || u.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and email are required")
	}
	created := h.store.Create(u)
	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) updateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var u User
	if err := c.Bind(&u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	updated, ok := h.store.Update(id, u)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user %d not found", id))
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) deleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if !h.store.Delete(id) {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user %d not found", id))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) getUserPosts(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if _, ok := h.store.Get(id); !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("user %d not found", id))
	}
	return c.JSON(http.StatusOK, []any{})
}

// --- Custom middleware ---

func requestID(next echo.HandlerFunc) echo.HandlerFunc {
	counter := 0
	return func(c echo.Context) error {
		counter++
		c.Response().Header().Set("X-Request-ID", fmt.Sprintf("req-%d", counter))
		return next(c)
	}
}

// --- Custom error handler ---

func errorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal server error"

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		msg = fmt.Sprintf("%v", he.Message)
	}

	if !c.Response().Committed {
		c.JSON(code, ErrorResponse{Error: msg})
	}
}

// --- Routes ---

func registerRoutes(e *echo.Echo, h *Handler) {
	// Health check
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "version": "1.0"})
	})

	// API v1 група — всі маршрути отримують /api/v1 префікс
	v1 := e.Group("/api/v1")

	// Users підгрупа з кастомним middleware
	users := v1.Group("/users", requestID)
	users.GET("", h.listUsers)
	users.POST("", h.createUser)
	users.GET("/:id", h.getUser)
	users.PUT("/:id", h.updateUser)
	users.DELETE("/:id", h.deleteUser)
	users.GET("/:id/posts", h.getUserPosts)
}

func main() {
	e := echo.New()
	e.HideBanner = true

	// Глобальні middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Кастомний error handler
	e.HTTPErrorHandler = errorHandler

	store := NewStore()
	handler := &Handler{store: store}
	registerRoutes(e, handler)

	e.Logger.Fatal(e.Start(":8080"))
}
