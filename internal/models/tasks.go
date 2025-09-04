package models

import (
    "time"
    "github.com/google/uuid"
)

type Task struct {
    ID              uuid.UUID `db:"id" json:"id"`
    Title           string    `db:"title" json:"title" binding:"required"`
    Description     string    `db:"description" json:"description"`
    Status          string    `db:"status" json:"status" binding:"required"`
    KanbanSpace     string    `db:"kanban_space" json:"kanban_space" binding:"required"` // "todo", "in_progress", "done", "review"
    Owner           string    `db:"owner" json:"owner" binding:"required"`
    AssignedTo      *string   `db:"assigned_to" json:"assigned_to,omitempty"` // Может быть nil
    Priority        string    `db:"priority" json:"priority"` // "low", "medium", "high", "urgent"
    DueDate         *time.Time `db:"due_date" json:"due_date,omitempty"`
    CreatedAt       time.Time `db:"created_at" json:"created_at"`
    UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// Допустимые значения для статусов
const (
    StatusTodo       = "todo"
    StatusInProgress = "in_progress"
    StatusReview     = "review"
    StatusDone       = "done"
)

// Допустимые значения для Kanban-пространств
const (
    SpaceBacklog     = "backlog"
    SpaceTodo        = "todo"
    SpaceInProgress  = "in_progress"
    SpaceReview      = "review"
    SpaceDone        = "done"
)

// Допустимые значения для приоритетов
const (
    PriorityLow      = "low"
    PriorityMedium   = "medium"
    PriorityHigh     = "high"
    PriorityUrgent   = "urgent"
)

// Validate проверяет валидность задачи
func (t *Task) Validate() error {
    if t.Title == "" {
        return &ValidationError{"title", "title is required"}
    }
    if t.Status == "" {
        return &ValidationError{"status", "status is required"}
    }
    if t.KanbanSpace == "" {
        return &ValidationError{"kanban_space", "kanban_space is required"}
    }
    if t.Owner == "" {
        return &ValidationError{"owner", "owner is required"}
    }
    
    // Валидация статуса
    switch t.Status {
    case StatusTodo, StatusInProgress, StatusReview, StatusDone:
        // OK
    default:
        return &ValidationError{"status", "invalid status"}
    }
    
    // Валидация Kanban-пространства
    switch t.KanbanSpace {
    case SpaceBacklog, SpaceTodo, SpaceInProgress, SpaceReview, SpaceDone:
        // OK
    default:
        return &ValidationError{"kanban_space", "invalid kanban space"}
    }
    
    // Валидация приоритета
    switch t.Priority {
    case "", PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
        
    default:
        return &ValidationError{"priority", "invalid priority"}
    }
    
    return nil
}

// ValidationError кастомная ошибка валидации
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Field + ": " + e.Message
}