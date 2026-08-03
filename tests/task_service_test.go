package service_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock Repository ─────────────────────────────────────────────────────────

type mockTaskRepo struct {
	mock.Mock
}

func (m *mockTaskRepo) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepo) FindAll(ctx context.Context, params dto.TaskQueryParams) ([]models.Task, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]models.Task), args.Get(1).(int64), args.Error(2)
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *mockTaskRepo) Update(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func newTaskService(repo *mockTaskRepo) service.TaskService {
	return service.NewTaskService(repo, nil) // nil redis = no caching
}

func sampleTask() *models.Task {
	id := uuid.New()
	now := time.Now()
	return &models.Task{
		ID:          id,
		Title:       "Test Task",
		Description: "Test description",
		Status:      models.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ─── CreateTask Tests ─────────────────────────────────────────────────────────

func TestCreateTask_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:       "Buy groceries",
		Description: "Milk, eggs, bread",
		Status:      "pending",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*models.Task")).Return(nil)

	result, err := svc.CreateTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Buy groceries", result.Title)
	assert.Equal(t, "pending", result.Status)
	repo.AssertExpectations(t)
}

func TestCreateTask_WithDueDate_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:   "Task with due date",
		DueDate: "2026-12-31",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*models.Task")).Return(nil)

	result, err := svc.CreateTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.DueDate)
	repo.AssertExpectations(t)
}

func TestCreateTask_InvalidDueDate(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	req := dto.CreateTaskRequest{
		Title:   "Task",
		DueDate: "not-a-date",
	}

	result, err := svc.CreateTask(ctx, req)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrInvalidDueDate)
	repo.AssertNotCalled(t, "Create")
}

func TestCreateTask_RepositoryError(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	req := dto.CreateTaskRequest{Title: "Task"}
	repo.On("Create", ctx, mock.AnythingOfType("*models.Task")).Return(errors.New("db error"))

	result, err := svc.CreateTask(ctx, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// ─── GetAllTasks Tests ────────────────────────────────────────────────────────

func TestGetAllTasks_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	tasks := []models.Task{*sampleTask(), *sampleTask()}
	params := dto.TaskQueryParams{Page: 1, Limit: 10}

	repo.On("FindAll", ctx, mock.AnythingOfType("dto.TaskQueryParams")).Return(tasks, int64(2), nil)

	result, err := svc.GetAllTasks(ctx, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Tasks, 2)
	assert.Equal(t, int64(2), result.Pagination.TotalTasks)
	assert.Equal(t, 1, result.Pagination.TotalPages)
	repo.AssertExpectations(t)
}

func TestGetAllTasks_DefaultPagination(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	// page=0, limit=0 should default to page=1, limit=10
	params := dto.TaskQueryParams{}
	repo.On("FindAll", ctx, mock.AnythingOfType("dto.TaskQueryParams")).Return([]models.Task{}, int64(0), nil)

	result, err := svc.GetAllTasks(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Pagination.CurrentPage)
	assert.Equal(t, 10, result.Pagination.Limit)
}

func TestGetAllTasks_RepositoryError(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	repo.On("FindAll", ctx, mock.AnythingOfType("dto.TaskQueryParams")).Return([]models.Task{}, int64(0), errors.New("db error"))

	result, err := svc.GetAllTasks(ctx, dto.TaskQueryParams{})

	assert.Nil(t, result)
	assert.Error(t, err)
}

// ─── GetTaskByID Tests ────────────────────────────────────────────────────────

func TestGetTaskByID_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	task := sampleTask()
	repo.On("FindByID", ctx, task.ID).Return(task, nil)

	result, err := svc.GetTaskByID(ctx, task.ID.String())

	assert.NoError(t, err)
	assert.Equal(t, task.ID.String(), result.ID)
	repo.AssertExpectations(t)
}

func TestGetTaskByID_NotFound(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	id := uuid.New()
	repo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := svc.GetTaskByID(ctx, id.String())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

func TestGetTaskByID_InvalidUUID(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	result, err := svc.GetTaskByID(ctx, "not-a-uuid")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
	repo.AssertNotCalled(t, "FindByID")
}

// ─── UpdateTask Tests ─────────────────────────────────────────────────────────

func TestUpdateTask_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	task := sampleTask()
	req := dto.UpdateTaskRequest{
		Title:  "Updated Title",
		Status: "completed",
	}

	repo.On("FindByID", ctx, task.ID).Return(task, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*models.Task")).Return(nil)

	result, err := svc.UpdateTask(ctx, task.ID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", result.Title)
	assert.Equal(t, "completed", result.Status)
	repo.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	id := uuid.New()
	repo.On("FindByID", ctx, id).Return(nil, nil)

	result, err := svc.UpdateTask(ctx, id.String(), dto.UpdateTaskRequest{Title: "New"})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

// ─── DeleteTask Tests ─────────────────────────────────────────────────────────

func TestDeleteTask_Success(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	id := uuid.New()
	repo.On("Delete", ctx, id).Return(nil)

	err := svc.DeleteTask(ctx, id.String())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteTask_NotFound(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	id := uuid.New()
	repo.On("Delete", ctx, id).Return(service.ErrTaskNotFound)

	err := svc.DeleteTask(ctx, id.String())

	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

func TestDeleteTask_InvalidUUID(t *testing.T) {
	repo := new(mockTaskRepo)
	svc := newTaskService(repo)
	ctx := context.Background()

	err := svc.DeleteTask(ctx, "bad-uuid")

	assert.ErrorIs(t, err, service.ErrTaskNotFound)
	repo.AssertNotCalled(t, "Delete")
}
