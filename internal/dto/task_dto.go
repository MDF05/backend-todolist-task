package dto

import (
	"time"
	"todo-api/internal/models"
)

// ─── Task Request DTOs ───────────────────────────────────────────────────────

// CreateTaskRequest is the request body for POST /tasks
type CreateTaskRequest struct {
	Title       string `json:"title"       validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=5000"`
	Status      string `json:"status"      validate:"omitempty,oneof=pending completed"`
	DueDate     string `json:"due_date"    validate:"omitempty,datetime=2006-01-02"`
}

// UpdateTaskRequest is the request body for PUT /tasks/:id
type UpdateTaskRequest struct {
	Title       string `json:"title"       validate:"omitempty,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=5000"`
	Status      string `json:"status"      validate:"omitempty,oneof=pending completed"`
	DueDate     string `json:"due_date"    validate:"omitempty,datetime=2006-01-02"`
}

// ─── Task Query Params ────────────────────────────────────────────────────────

// TaskQueryParams holds query parameters for GET /tasks
type TaskQueryParams struct {
	Status string `form:"status"  validate:"omitempty,oneof=pending completed"`
	Page   int    `form:"page"    validate:"omitempty,min=1"`
	Limit  int    `form:"limit"   validate:"omitempty,min=1,max=100"`
	Search string `form:"search"  validate:"omitempty,max=255"`
}

// ─── Task Response DTOs ──────────────────────────────────────────────────────

// TaskResponse is the standard task representation in API responses
type TaskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalTasks  int64 `json:"total_tasks"`
	Limit       int   `json:"limit"`
}

// TaskListResponse is the paginated list response for GET /tasks
type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Pagination PaginationMeta `json:"pagination"`
}

// ─── Auth DTOs ───────────────────────────────────────────────────────────────

// RegisterRequest is the body for POST /auth/register
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// LoginRequest is the body for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned on successful login
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// UserResponse is the public representation of a user
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ─── Generic API Response ─────────────────────────────────────────────────────

// SuccessResponse wraps a successful response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse wraps an error response
type ErrorResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Errors  []FieldError  `json:"errors,omitempty"`
}

// FieldError represents a validation error for a single field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ─── Mappers ─────────────────────────────────────────────────────────────────

// FromTask converts a Task model to a TaskResponse DTO
func FromTask(task *models.Task) TaskResponse {
	resp := TaskResponse{
		ID:          task.ID.String(),
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}
	if task.DueDate != nil {
		d := task.DueDate.Format("2006-01-02")
		resp.DueDate = &d
	}
	return resp
}

// FromTasks converts a slice of Task models to TaskResponse DTOs
func FromTasks(tasks []models.Task) []TaskResponse {
	result := make([]TaskResponse, 0, len(tasks))
	for i := range tasks {
		result = append(result, FromTask(&tasks[i]))
	}
	return result
}
