-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'todo',
    kanban_space VARCHAR(50) NOT NULL DEFAULT 'backlog',
    owner VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    priority VARCHAR(20) DEFAULT 'medium',
    due_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

--индексы для оптимизации
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_kanban_space ON tasks(kanban_space);
CREATE INDEX idx_tasks_owner ON tasks(owner);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- ограничения
ALTER TABLE tasks ADD CONSTRAINT chk_status CHECK (status IN ('todo', 'in_progress', 'review', 'done'));
ALTER TABLE tasks ADD CONSTRAINT chk_kanban_space CHECK (kanban_space IN ('backlog', 'todo', 'in_progress', 'review', 'done'));
ALTER TABLE tasks ADD CONSTRAINT chk_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tasks;
-- +goose StatementEnd