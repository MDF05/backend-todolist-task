package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus defines valid status values for a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusCompleted TaskStatus = "completed"
)

// Task represents the tasks table in PostgreSQL
type Task struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null"                     json:"title"`
	Description string         `gorm:"type:text"                                      json:"description"`
	Status      TaskStatus     `gorm:"type:varchar(20);not null;default:'pending'"    json:"status"`
	DueDate     *time.Time     `gorm:"type:date"                                      json:"due_date"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

// BeforeCreate hook: generate UUID if not set
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Task model
func (Task) TableName() string {
	return "tasks"
}
