package repository

import (
    "context"
    "errors"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/TrueSmartcomm/backend/internal/models"
)

type TaskRepository struct {
    DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
    return &TaskRepository{DB: db}
}

// CreateTask создает новую задачу
func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
    if task.ID == uuid.Nil {
        task.ID = uuid.New()
    }
    
    // Устанавливаем дефолтные значения
    if task.Status == "" {
        task.Status = models.StatusTodo
    }
    if task.KanbanSpace == "" {
        task.KanbanSpace = models.SpaceBacklog
    }
    if task.Priority == "" {
        task.Priority = models.PriorityMedium
    }
    
    query := `INSERT INTO tasks (id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()) RETURNING created_at, updated_at`
    
    err := r.DB.QueryRow(ctx, query, 
        task.ID, task.Title, task.Description, task.Status, task.KanbanSpace, 
        task.Owner, task.AssignedTo, task.Priority, task.DueDate).
        Scan(&task.CreatedAt, &task.UpdatedAt)
    
    if err != nil {
        return err
    }
    
    return nil
}

// GetTaskByID получает задачу по ID
func (r *TaskRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
    var task models.Task
    query := `SELECT id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at 
              FROM tasks WHERE id = $1`
    
    err := r.DB.QueryRow(ctx, query, id).
        Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.KanbanSpace, 
             &task.Owner, &task.AssignedTo, &task.Priority, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("task not found")
        }
        return nil, err
    }
    
    return &task, nil
}

// UpdateTask обновляет задачу
func (r *TaskRepository) UpdateTask(ctx context.Context, task *models.Task) error {
    query := `UPDATE tasks SET title=$1, description=$2, status=$3, kanban_space=$4, owner=$5, 
              assigned_to=$6, priority=$7, due_date=$8, updated_at=now() 
              WHERE id=$9 RETURNING updated_at`
    
    err := r.DB.QueryRow(ctx, query, 
        task.Title, task.Description, task.Status, task.KanbanSpace, 
        task.Owner, task.AssignedTo, task.Priority, task.DueDate, task.ID).
        Scan(&task.UpdatedAt)
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return errors.New("task not found")
        }
        return err
    }
    
    return nil
}

// DeleteTask удаляет задачу
func (r *TaskRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
    result, err := r.DB.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
    if err != nil {
        return err
    }
    
    if result.RowsAffected() == 0 {
        return errors.New("task not found")
    }
    
    return nil
}

// GetAllTasks получает все задачи (по фильтрам)
func (r *TaskRepository) GetAllTasks(ctx context.Context, space string, status string) ([]models.Task, error) {
    var query string
    var args []interface{}
    
    if space != "" && status != "" {
        query = `SELECT id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at 
                 FROM tasks WHERE kanban_space=$1 AND status=$2 ORDER BY created_at DESC`
        args = []interface{}{space, status}
    } else if space != "" {
        query = `SELECT id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at 
                 FROM tasks WHERE kanban_space=$1 ORDER BY created_at DESC`
        args = []interface{}{space}
    } else if status != "" {
        query = `SELECT id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at 
                 FROM tasks WHERE status=$1 ORDER BY created_at DESC`
        args = []interface{}{status}
    } else {
        query = `SELECT id, title, description, status, kanban_space, owner, assigned_to, priority, due_date, created_at, updated_at 
                 FROM tasks ORDER BY created_at DESC`
    }
    
    rows, err := r.DB.Query(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var task models.Task
        err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.KanbanSpace, 
                        &task.Owner, &task.AssignedTo, &task.Priority, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, task)
    }
    
    if err = rows.Err(); err != nil {
        return nil, err
    }
    
    return tasks, nil
}

// MoveTaskToSpace перемещает задачу в другое Kanban-пространство
func (r *TaskRepository) MoveTaskToSpace(ctx context.Context, id uuid.UUID, space string, status string) error {
    query := `UPDATE tasks SET kanban_space=$1, status=$2, updated_at=now() WHERE id=$3`
    
    result, err := r.DB.Exec(ctx, query, space, status, id)
    if err != nil {
        return err
    }
    
    if result.RowsAffected() == 0 {
        return errors.New("task not found")
    }
    
    return nil
}