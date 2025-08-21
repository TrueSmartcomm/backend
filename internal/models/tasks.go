package models

import "time"

type Task struct {
    ID          int       `db:"id" json:"id"`
    Title       string    `db:"title" json:"title"`
    Description string    `db:"description" json:"description"`
    Status      string    `db:"status" json:"status"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
