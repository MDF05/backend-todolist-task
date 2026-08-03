package repository

import (
	"context"
	"todo-api/internal/dto"
	"todo-api/internal/models"

	"github.com/google/uuid"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindAll(ctx context.Context, params dto.TaskQueryParams) ([]models.Task, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}
