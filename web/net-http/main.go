package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
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

// --- In-memory store ---

type Store struct {
	mu      sync.RWMutex
	users   map[int]User
	nextID  int
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

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
	return json.NewDecoder(r.Body).Decode(v)
}

func pathID(r *http.Request, param string) (int, error) {
	return strconv.Atoi(r.PathValue(param))
}

// --- Middleware ---

type Middleware func(http.Handler) http.Handler

// Chain застосовує middleware у прямому порядку (перший — найзовніший)
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start),
		)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic", "error", err)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func BasicAuth(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || u != username || p != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter перехоплює статус код
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// --- Handlers ---

type Handler struct {
	store *Store
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetAll())
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("user %d not found", id))
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := decodeJSON(w, r, &u); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if u.Name == "" || u.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}
	created := h.store.Create(u)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var u User
	if err := decodeJSON(w, r, &u); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, ok := h.store.Update(id, u)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("user %d not found", id))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.store.Delete(id) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("user %d not found", id))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getUserPosts(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if _, ok := h.store.Get(id); !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("user %d not found", id))
		return
	}
	// Демо: повертаємо порожній список
	writeJSON(w, http.StatusOK, []any{})
}

// --- Router ---

func newRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0"})
	})

	// Users CRUD
	mux.HandleFunc("GET /api/users", h.listUsers)
	mux.HandleFunc("POST /api/users", h.createUser)
	mux.HandleFunc("GET /api/users/{id}", h.getUser)
	mux.HandleFunc("PUT /api/users/{id}", h.updateUser)
	mux.HandleFunc("DELETE /api/users/{id}", h.deleteUser)

	// Вкладені ресурси
	mux.HandleFunc("GET /api/users/{id}/posts", h.getUserPosts)

	return Chain(mux, Logging, Recovery)
}

func main() {
	store := NewStore()
	handler := &Handler{store: store}
	router := newRouter(handler)

	addr := ":8080"
	slog.Info("server starting", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
