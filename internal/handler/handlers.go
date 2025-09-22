package handlers

import (
	"context"
	"net/http"

	"github.com/TrueSmartcomm/backend/internal/models"
	"github.com/TrueSmartcomm/backend/internal/repository"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	repo *repository.TaskRepository
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// POST /tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if task.Status == "" {
		task.Status = "todo"
	}

	if err := h.repo.CreateTask(context.Background(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

// GET /tasks
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Query("id")
	id := uuid.MustParse(idStr)

	task, err := h.repo.GetTaskByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// PUT /tasks
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateTask(context.Background(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// DELETE
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Query("id")
	id := uuid.MustParse(idStr)

	if err := h.repo.DeleteTask(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *TaskHandler) MoveTask(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	var req struct {
		Space  string `json:"space" binding:"required"`
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация значений
	validSpaces := map[string]bool{
		models.SpaceBacklog: true, models.SpaceTodo: true, models.SpaceInProgress: true,
		models.SpaceReview: true, models.SpaceDone: true,
	}
	validStatuses := map[string]bool{
		models.StatusTodo: true, models.StatusInProgress: true,
		models.StatusReview: true, models.StatusDone: true,
	}

	if !validSpaces[req.Space] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid space"})
		return
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err := h.repo.MoveTaskToSpace(context.Background(), id, req.Space, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mesage": "moved", "space": req.Space, "status": req.Status})
}

// AddTaskDependency добавляет зависимость между задачами
func (h *TaskHandler) AddTaskDependency(c *gin.Context) {
	var req struct {
		TaskID          string `json:"task_id" binding:"required"`
		DependentTaskID string `json:"dependent_task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id format"})
		return
	}

	dependentTaskID, err := uuid.Parse(req.DependentTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dependent_task_id format"})
		return
	}

	// Проверка существования обеих задач
	_, err = h.repo.GetTaskByID(c, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	_, err = h.repo.GetTaskByID(c, dependentTaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dependent task not found"})
		return
	}

	if err := h.repo.AddDependency(c, taskID, dependentTaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "dependency added"})
}

func (h *TaskHandler) RemoveTaskDependency(c *gin.Context) {
	var req struct {
		TaskID          string `json:"task_id" binding:"required"`
		DependentTaskID string `json:"dependent_task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id format"})
		return
	}

	dependentTaskID, err := uuid.Parse(req.DependentTaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dependent_task_id format"})
		return
	}

	if err := h.repo.RemoveDependency(c, taskID, dependentTaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "dependency removed"})
}

func (h *TaskHandler) GetTaskWithDependencies(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	task, err := h.repo.GetTaskByIDWithDependencies(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
