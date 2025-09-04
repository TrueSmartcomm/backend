package handler

import (
    "context"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/TrueSmartcomm/backend/internal/models"
    "github.com/TrueSmartcomm/backend/internal/repository"
)

type TaskHandler struct {
    repo *repository.TaskRepository
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
    return &TaskHandler{repo: repo}
}

// CreateTask создает новую задачу
func (h *TaskHandler) CreateTask(c *gin.Context) {
    var task models.Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Валидация
    if err := task.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.repo.CreateTask(context.Background(), &task); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, task)
}

// GetTask получает задачу по ID из query параметра
func (h *TaskHandler) GetTask(c *gin.Context) {
    idStr := c.Query("id")
    if idStr == "" {
        // Если ID не указан, возвращаем список задач с фильтрами
        space := c.Query("space")
        status := c.Query("status")
        
        tasks, err := h.repo.GetAllTasks(context.Background(), space, status)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, tasks)
        return
    }

    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
        return
    }

    task, err := h.repo.GetTaskByID(context.Background(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
        return
    }
    c.JSON(http.StatusOK, task)
}

// UpdateTask обновляет задачу
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    var task models.Task
    if err := c.ShouldBindJSON(&task); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Проверяем, что ID передан
    if task.ID == uuid.Nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "task id is required"})
        return
    }

    // Валидация
    if err := task.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.repo.UpdateTask(context.Background(), &task); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, task)
}

// DeleteTask удаляет задачу
func (h *TaskHandler) DeleteTask(c *gin.Context) {
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

    if err := h.repo.DeleteTask(context.Background(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// MoveTask перемещает задачу в другое Kanban-пространство
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