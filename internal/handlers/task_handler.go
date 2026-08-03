package handlers

import (
	"errors"
	"net/http"
	"todo-api/internal/dto"
	"todo-api/internal/service"

	"github.com/gin-gonic/gin"
)

// TaskHandler handles all task-related HTTP requests
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler creates a new TaskHandler
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// CreateTask godoc
// @Summary      Create a new task
// @Description  Create a new task for the authenticated user
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateTaskRequest true "Task data"
// @Success      201  {object}  dto.SuccessResponse{data=dto.TaskResponse}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDueDate) {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "Failed to create task")
		return
	}

	respondSuccess(c, http.StatusCreated, "Task created successfully", task)
}

// GetAllTasks godoc
// @Summary      Get all tasks
// @Description  Get a paginated list of tasks with optional filtering and search
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "Filter by status (pending/completed)"
// @Param        page query int false "Page number (default 1)"
// @Param        limit query int false "Items per page (default 10)"
// @Param        search query string false "Search query"
// @Success      200  {object}  dto.SuccessResponse{data=dto.TaskListResponse}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	var params dto.TaskQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		respondValidationError(c, err)
		return
	}

	if err := validate.Struct(params); err != nil {
		respondValidationError(c, err)
		return
	}

	result, err := h.taskService.GetAllTasks(c.Request.Context(), params)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetTaskByID godoc
// @Summary      Get a task by ID
// @Description  Get details of a specific task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200  {object}  dto.SuccessResponse{data=dto.TaskResponse}
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")

	task, err := h.taskService.GetTaskByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			respondError(c, http.StatusNotFound, "Task not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   task,
	})
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Update details of an existing task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Param        request body dto.UpdateTaskRequest true "Update data"
// @Success      200  {object}  dto.SuccessResponse{data=dto.TaskResponse}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(c, err)
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			respondError(c, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrInvalidDueDate):
			respondError(c, http.StatusBadRequest, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "Failed to update task")
		}
		return
	}

	respondSuccess(c, http.StatusOK, "Task updated successfully", task)
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Delete an existing task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200  {object}  dto.SuccessResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	err := h.taskService.DeleteTask(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			respondError(c, http.StatusNotFound, "Task not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Task deleted successfully",
	})
}
