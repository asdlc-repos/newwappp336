package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/asdlc-repos/newwappp336/todo-api/internal/auth"
	"github.com/asdlc-repos/newwappp336/todo-api/internal/models"
	"github.com/asdlc-repos/newwappp336/todo-api/internal/store"
)

// API bundles the handler dependencies.
type API struct {
	Store *store.Store
}

// New returns an API bound to the given store.
func New(s *store.Store) *API {
	return &API{Store: s}
}

// Register wires every route to the provided mux.
// Authenticated routes are protected via auth.Middleware.
func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/auth/signup", a.handleSignup)
	mux.HandleFunc("/auth/login", a.handleLogin)

	mux.Handle("/categories", auth.Middleware(http.HandlerFunc(a.handleCategories)))
	mux.Handle("/categories/", auth.Middleware(http.HandlerFunc(a.handleCategoryByID)))
	mux.Handle("/tasks", auth.Middleware(http.HandlerFunc(a.handleTasks)))
	mux.Handle("/tasks/", auth.Middleware(http.HandlerFunc(a.handleTaskByID)))
}

// -------------------- Health --------------------

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// -------------------- Auth --------------------

func (a *API) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req models.AuthRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	u, err := a.Store.CreateUser(email, hash)
	if err != nil {
		if errors.Is(err, store.ErrUserExists) {
			writeError(w, http.StatusConflict, "user already exists")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := auth.IssueToken(u.ID)
	if err != nil {
		log.Printf("token issue error: %v", err)
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	writeJSON(w, http.StatusCreated, models.AuthResponse{Token: token, UserID: u.ID})
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req models.AuthRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	u, err := a.Store.GetUserByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := auth.IssueToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	writeJSON(w, http.StatusOK, models.AuthResponse{Token: token, UserID: u.ID})
}

// -------------------- Categories --------------------

func (a *API) handleCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		cats := a.Store.ListCategories(userID)
		writeJSON(w, http.StatusOK, cats)
	case http.MethodPost:
		var in models.CategoryInput
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		c, err := a.Store.CreateCategory(userID, in.Name)
		if err != nil {
			if errors.Is(err, store.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, "name is required")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, c)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/categories/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := a.Store.DeleteCategory(userID, id); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeMethodNotAllowed(w)
	}
}

// -------------------- Tasks --------------------

func (a *API) handleTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		categoryFilter := r.URL.Query().Get("categoryId")
		tasks := a.Store.ListTasks(userID, categoryFilter)
		writeJSON(w, http.StatusOK, tasks)
	case http.MethodPost:
		var in models.TaskInput
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		t, err := a.Store.CreateTask(userID, in)
		if err != nil {
			if errors.Is(err, store.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, "invalid task input")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, "/tasks/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	parts := strings.Split(rest, "/")
	id := parts[0]

	// /tasks/{id}/complete
	if len(parts) == 2 && parts[1] == "complete" {
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}
		completedStr := r.URL.Query().Get("completed")
		completed := true
		if completedStr != "" {
			v, err := strconv.ParseBool(completedStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid completed value")
				return
			}
			completed = v
		}
		t, err := a.Store.SetTaskCompleted(userID, id, completed)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, t)
		return
	}

	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var in models.TaskInput
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		t, err := a.Store.UpdateTask(userID, id, in)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, t)
	case http.MethodDelete:
		if err := a.Store.DeleteTask(userID, id); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeMethodNotAllowed(w)
	}
}

// -------------------- Helpers --------------------

func decodeJSON(r *http.Request, v interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, store.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, store.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	default:
		log.Printf("store error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
