package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/TrueSmartcomm/backend/internal/models"
)

type TaskRepository struct {
    DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
    return &TaskRepository{DB: db}
}

// Создать задачу
func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
    query := `INSERT INTO tasks (title, description, status, created_at, updated_at) 
              VALUES ($1, $2, $3, now(), now()) RETURNING id, created_at, updated_at`
    return r.DB.QueryRow(ctx, query, task.Title, task.Description, task.Status).
        Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

// Получить задачу по ID
func (r *TaskRepository) GetTaskByID(ctx context.Context, id int) (*models.Task, error) {
    var task models.Task
    query := `SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1`
    err := r.DB.QueryRow(ctx, query, id).
        Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &task, nil
}

// Обновить задачу
func (r *TaskRepository) UpdateTask(ctx context.Context, task *models.Task) error {
    query := `UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=now() WHERE id=$4 RETURNING updated_at`
    return r.DB.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ID).
        Scan(&task.UpdatedAt)
}

// Удалить задачу
func (r *TaskRepository) DeleteTask(ctx context.Context, id int) error {
    _, err := r.DB.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, id)
    return err
}