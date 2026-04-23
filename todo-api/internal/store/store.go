package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/asdlc-repos/newwappp336/todo-api/internal/models"
)

// Errors returned by the Store.
var (
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

// Store is a thread-safe in-memory store for users, categories, and tasks.
type Store struct {
	mu            sync.RWMutex
	users         map[string]*models.User // userID -> User
	usersByEmail  map[string]*models.User // lowercase email -> User
	categories    map[string]*models.Category
	tasks         map[string]*models.Task
}

// New creates an empty Store.
func New() *Store {
	return &Store{
		users:        make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		categories:   make(map[string]*models.Category),
		tasks:        make(map[string]*models.Task),
	}
}

// -------------------- Users --------------------

// CreateUser inserts a new user. Returns ErrUserExists if email already registered.
func (s *Store) CreateUser(email string, passwordHash []byte) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usersByEmail[email]; ok {
		return nil, ErrUserExists
	}

	u := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
	}
	s.users[u.ID] = u
	s.usersByEmail[email] = u
	return u, nil
}

// GetUserByEmail looks up a user by email (case-insensitive).
func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// GetUserByID returns the user with the given ID.
func (s *Store) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// -------------------- Categories --------------------

// ListCategories returns the categories owned by userID.
func (s *Store) ListCategories(userID string) []*models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Category, 0)
	for _, c := range s.categories {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out
}

// CreateCategory creates a new category for userID.
func (s *Store) CreateCategory(userID, name string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := &models.Category{
		ID:     uuid.NewString(),
		UserID: userID,
		Name:   name,
	}
	s.categories[c.ID] = c
	return c, nil
}

// DeleteCategory removes a category and nulls categoryId on that user's tasks.
// Returns ErrNotFound if the category doesn't exist, ErrForbidden if owned by a different user.
func (s *Store) DeleteCategory(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.categories[id]
	if !ok {
		return ErrNotFound
	}
	if c.UserID != userID {
		return ErrForbidden
	}

	delete(s.categories, id)

	for _, t := range s.tasks {
		if t.UserID == userID && t.CategoryID == id {
			t.CategoryID = ""
		}
	}
	return nil
}

// categoryExistsForUser returns true iff category id exists and is owned by userID.
// Caller must hold at least read lock.
func (s *Store) categoryExistsForUser(userID, id string) bool {
	if id == "" {
		return true
	}
	c, ok := s.categories[id]
	return ok && c.UserID == userID
}

// -------------------- Tasks --------------------

// ListTasks returns the tasks for userID, optionally filtered by categoryID.
// An empty categoryFilter returns all tasks for the user.
func (s *Store) ListTasks(userID string, categoryFilter string) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Task, 0)
	for _, t := range s.tasks {
		if t.UserID != userID {
			continue
		}
		if categoryFilter != "" && t.CategoryID != categoryFilter {
			continue
		}
		out = append(out, t)
	}
	return out
}

// CreateTask creates a new task for userID.
func (s *Store) CreateTask(userID string, input models.TaskInput) (*models.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	categoryID := ""
	if input.CategoryID != nil {
		categoryID = *input.CategoryID
	}
	if categoryID != "" && !s.categoryExistsForUser(userID, categoryID) {
		return nil, ErrInvalidInput
	}

	completed := false
	if input.Completed != nil {
		completed = *input.Completed
	}

	t := &models.Task{
		ID:         uuid.NewString(),
		UserID:     userID,
		Title:      title,
		DueDate:    input.DueDate,
		CategoryID: categoryID,
		Completed:  completed,
		CreatedAt:  time.Now().UTC(),
	}
	s.tasks[t.ID] = t
	return t, nil
}

// GetTask returns a single task.
func (s *Store) GetTask(userID, id string) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	if t.UserID != userID {
		return nil, ErrForbidden
	}
	return t, nil
}

// UpdateTask mutates an existing task's fields using values supplied in input.
// All fields present in the input replace the stored task's values; the title
// is required and must be non-empty.
func (s *Store) UpdateTask(userID, id string, input models.TaskInput) (*models.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	if t.UserID != userID {
		return nil, ErrForbidden
	}

	categoryID := ""
	if input.CategoryID != nil {
		categoryID = *input.CategoryID
	}
	if categoryID != "" && !s.categoryExistsForUser(userID, categoryID) {
		return nil, ErrInvalidInput
	}

	t.Title = title
	t.DueDate = input.DueDate
	t.CategoryID = categoryID
	if input.Completed != nil {
		t.Completed = *input.Completed
	}
	return t, nil
}

// DeleteTask removes a task.
func (s *Store) DeleteTask(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}
	if t.UserID != userID {
		return ErrForbidden
	}
	delete(s.tasks, id)
	return nil
}

// SetTaskCompleted updates only the completion flag.
func (s *Store) SetTaskCompleted(userID, id string, completed bool) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	if t.UserID != userID {
		return nil, ErrForbidden
	}
	t.Completed = completed
	return t, nil
}
