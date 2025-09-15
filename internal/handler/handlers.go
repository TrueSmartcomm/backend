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
