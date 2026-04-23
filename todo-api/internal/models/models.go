package models

import "time"

// User represents an authenticated application user.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
}

// Category is a per-user grouping for tasks.
type Category struct {
	ID     string `json:"id"`
	UserID string `json:"-"`
	Name   string `json:"name"`
}

// Task is a todo item owned by a user and optionally assigned to a category.
type Task struct {
	ID         string     `json:"id"`
	UserID     string     `json:"-"`
	Title      string     `json:"title"`
	DueDate    *time.Time `json:"dueDate"`
	CategoryID string     `json:"categoryId"`
	Completed  bool       `json:"completed"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// AuthRequest is the request body for /auth/signup and /auth/login.
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response body for successful auth operations.
type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
}

// TaskInput is the request body for task create/update.
type TaskInput struct {
	Title      string     `json:"title"`
	DueDate    *time.Time `json:"dueDate"`
	CategoryID *string    `json:"categoryId"`
	Completed  *bool      `json:"completed"`
}

// CategoryInput is the request body for category create.
type CategoryInput struct {
	Name string `json:"name"`
}
