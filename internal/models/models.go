package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	GitHubID    int64     `json:"github_id" gorm:"uniqueIndex"`
	Username    string    `json:"username" gorm:"uniqueIndex"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	AccessToken string    `json:"-" gorm:"column:access_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Workflow represents a workflow definition
type Workflow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	YAMLContent string    `json:"yaml_content"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	Runs        []Run     `json:"runs,omitempty" gorm:"foreignKey:WorkflowID"`
}

// Run represents a workflow execution
type Run struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	WorkflowID uint      `json:"workflow_id"`
	UserID     uint      `json:"user_id"`
	Status     string    `json:"status"` // pending, running, success, failed, cancelled
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Workflow   Workflow  `json:"workflow" gorm:"foreignKey:WorkflowID"`
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	Jobs       []Job     `json:"jobs,omitempty" gorm:"foreignKey:RunID"`
}

// Job represents a job within a workflow run
type Job struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RunID     uint      `json:"run_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // pending, running, success, failed, skipped
	RunnerID  string    `json:"runner_id"`
	StartedAt *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Run       Run       `json:"run" gorm:"foreignKey:RunID"`
	Steps     []Step    `json:"steps,omitempty" gorm:"foreignKey:JobID"`
}

// Step represents a step within a job
type Step struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	JobID     uint       `json:"job_id"`
	Name      string     `json:"name"`
	Command   string     `json:"command"`
	Status    string     `json:"status"` // pending, running, success, failed, skipped
	StartedAt *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Job       Job        `json:"job" gorm:"foreignKey:JobID"`
	Logs      []Log      `json:"logs,omitempty" gorm:"foreignKey:StepID"`
}

// Log represents log entries for steps
type Log struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	StepID    uint      `json:"step_id"`
	Content   string    `json:"content"`
	Level     string    `json:"level"` // info, warn, error, debug
	Timestamp time.Time `json:"timestamp"`
	Step      Step      `json:"step" gorm:"foreignKey:StepID"`
}

// Runner represents a workflow runner instance
type Runner struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // online, offline, busy
	LastSeen  time.Time `json:"last_seen"`
	Version   string    `json:"version"`
	Tags      string    `json:"tags"` // JSON array of tags
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}