package repository

import (
	"context"
	"errors"
	"math"
	"todo-api/internal/dto"
	"todo-api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type taskRepositoryImpl struct {
	db *gorm.DB
}

// NewTaskRepository creates a new TaskRepository backed by GORM
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

// Create inserts a new task into the database
func (r *taskRepositoryImpl) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// FindAll retrieves tasks with optional filters, pagination, and search
func (r *taskRepositoryImpl) FindAll(ctx context.Context, params dto.TaskQueryParams) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Task{})

	// Filter by status
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Full-text search on title and description
	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	// Count total matching records (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Fetch paginated results, ordered by newest first
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// FindByID retrieves a single task by its UUID
func (r *taskRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found — caller decides how to handle
		}
		return nil, err
	}
	return &task, nil
}

// Update saves changes to an existing task
func (r *taskRepositoryImpl) Update(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete performs a soft delete on the task
func (r *taskRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CalcTotalPages calculates total pages from total records and page size
func CalcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
