package models

import (
	"time"
)

type Agent struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	URL         string    `json:"url" db:"url"`
	APIKey      string    `json:"api_key,omitempty" db:"api_key"`
	LastSeenAt  time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description" db:"description"`
}

type CreateNodeRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

type CreateNodeResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	APIKey      string `json:"api_key"`
	CreatedAt   string `json:"created_at"`
}

type RegisterRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	URL  string `json:"url" validate:"required,url"`
}

type RegisterResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
}

type HeartbeatRequest struct {
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type HeartbeatResponse struct {
	Status string `json:"status"`
	Tasks  []Task `json:"tasks"`
}

type Task struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}
