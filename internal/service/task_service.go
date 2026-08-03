package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// ─── Errors ──────────────────────────────────────────────────────────────────

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidStatus   = errors.New("invalid status value")
	ErrInvalidDueDate  = errors.New("invalid due_date format, expected YYYY-MM-DD")
)

// ─── Interface ───────────────────────────────────────────────────────────────

// TaskService defines all business logic operations for tasks
type TaskService interface {
	CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetAllTasks(ctx context.Context, params dto.TaskQueryParams) (*dto.TaskListResponse, error)
	GetTaskByID(ctx context.Context, id string) (*dto.TaskResponse, error)
	UpdateTask(ctx context.Context, id string, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	DeleteTask(ctx context.Context, id string) error
}

// ─── Implementation ──────────────────────────────────────────────────────────

type taskServiceImpl struct {
	repo        repository.TaskRepository
	redisClient *redis.Client
	cacheTTL    time.Duration
}

// NewTaskService creates a new TaskService
func NewTaskService(repo repository.TaskRepository, redisClient *redis.Client) TaskService {
	return &taskServiceImpl{
		repo:        repo,
		redisClient: redisClient,
		cacheTTL:    5 * time.Minute,
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func (s *taskServiceImpl) CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusPending,
	}

	if req.Status != "" {
		task.Status = models.TaskStatus(req.Status)
	}

	if req.DueDate != "" {
		d, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, ErrInvalidDueDate
		}
		task.DueDate = &d
	}

	if err := s.repo.Create(ctx, task); err != nil {
		log.Error().Err(err).Msg("failed to create task")
		return nil, err
	}

	// Invalidate list cache asynchronously (concurrent execution)
	go s.invalidateListCache()

	resp := dto.FromTask(task)
	return &resp, nil
}

// ─── Get All ─────────────────────────────────────────────────────────────────

func (s *taskServiceImpl) GetAllTasks(ctx context.Context, params dto.TaskQueryParams) (*dto.TaskListResponse, error) {
	// Default pagination values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}

	cacheKey := s.buildListCacheKey(params)

	// Try Redis cache first
	if s.redisClient != nil {
		if cached, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil {
			var result dto.TaskListResponse
			if json.Unmarshal([]byte(cached), &result) == nil {
				log.Debug().Str("cache_key", cacheKey).Msg("cache hit: task list")
				return &result, nil
			}
		}
	}

	// ── Concurrent execution: fetch tasks and count in parallel ──────────────
	type fetchResult struct {
		tasks []models.Task
		total int64
		err   error
	}

	resultCh := make(chan fetchResult, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		tasks, total, err := s.repo.FindAll(ctx, params)
		resultCh <- fetchResult{tasks, total, err}
	}()

	wg.Wait()
	close(resultCh)

	res := <-resultCh
	if res.err != nil {
		log.Error().Err(res.err).Msg("failed to fetch tasks")
		return nil, res.err
	}

	totalPages := repository.CalcTotalPages(res.total, params.Limit)

	response := &dto.TaskListResponse{
		Tasks: dto.FromTasks(res.tasks),
		Pagination: dto.PaginationMeta{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
			TotalTasks:  res.total,
			Limit:       params.Limit,
		},
	}

	// Cache the result asynchronously
	if s.redisClient != nil {
		go func() {
			data, _ := json.Marshal(response)
			bgCtx := context.Background()
			if err := s.redisClient.Set(bgCtx, cacheKey, data, s.cacheTTL).Err(); err != nil {
				log.Warn().Err(err).Str("key", cacheKey).Msg("failed to cache task list")
			}
		}()
	}

	return response, nil
}

// ─── Get By ID ───────────────────────────────────────────────────────────────

func (s *taskServiceImpl) GetTaskByID(ctx context.Context, id string) (*dto.TaskResponse, error) {
	taskID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	cacheKey := fmt.Sprintf("task:%s", id)

	// Try Redis cache first
	if s.redisClient != nil {
		if cached, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil {
			var result dto.TaskResponse
			if json.Unmarshal([]byte(cached), &result) == nil {
				log.Debug().Str("task_id", id).Msg("cache hit: task by id")
				return &result, nil
			}
		}
	}

	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Str("task_id", id).Msg("failed to get task")
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	resp := dto.FromTask(task)

	// Cache it asynchronously
	if s.redisClient != nil {
		go func() {
			data, _ := json.Marshal(resp)
			bgCtx := context.Background()
			if err := s.redisClient.Set(bgCtx, cacheKey, data, s.cacheTTL).Err(); err != nil {
				log.Warn().Err(err).Str("task_id", id).Msg("failed to cache task")
			}
		}()
	}

	return &resp, nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *taskServiceImpl) UpdateTask(ctx context.Context, id string, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	taskID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		log.Error().Err(err).Str("task_id", id).Msg("error finding task for update")
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	// Patch only provided fields
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = models.TaskStatus(req.Status)
	}
	if req.DueDate != "" {
		d, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, ErrInvalidDueDate
		}
		task.DueDate = &d
	}

	if err := s.repo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", id).Msg("failed to update task")
		return nil, err
	}

	// Invalidate caches concurrently
	go func() {
		bgCtx := context.Background()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.invalidateTaskCache(bgCtx, id)
		}()
		go func() {
			defer wg.Done()
			s.invalidateListCache()
		}()
		wg.Wait()
	}()

	resp := dto.FromTask(task)
	return &resp, nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *taskServiceImpl) DeleteTask(ctx context.Context, id string) error {
	taskID, err := uuid.Parse(id)
	if err != nil {
		return ErrTaskNotFound
	}

	err = s.repo.Delete(ctx, taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return ErrTaskNotFound
		}
		log.Error().Err(err).Str("task_id", id).Msg("failed to delete task")
		return err
	}

	// Invalidate caches concurrently
	go func() {
		bgCtx := context.Background()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.invalidateTaskCache(bgCtx, id)
		}()
		go func() {
			defer wg.Done()
			s.invalidateListCache()
		}()
		wg.Wait()
	}()

	return nil
}

// ─── Cache Helpers ────────────────────────────────────────────────────────────

func (s *taskServiceImpl) buildListCacheKey(params dto.TaskQueryParams) string {
	return fmt.Sprintf("tasks:list:status=%s:page=%d:limit=%d:search=%s",
		params.Status, params.Page, params.Limit, params.Search)
}

func (s *taskServiceImpl) invalidateListCache() {
	if s.redisClient == nil {
		return
	}
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, "tasks:list:*").Result()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get list cache keys")
		return
	}
	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Warn().Err(err).Msg("failed to invalidate list cache")
		}
	}
}

func (s *taskServiceImpl) invalidateTaskCache(ctx context.Context, id string) {
	if s.redisClient == nil {
		return
	}
	key := fmt.Sprintf("task:%s", id)
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("failed to invalidate task cache")
	}
}
