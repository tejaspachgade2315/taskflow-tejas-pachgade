package store

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWithPassword struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	ProjectID    string     `json:"project_id"`
	AssigneeID   *string    `json:"assignee_id,omitempty"`
	AssigneeName *string    `json:"assignee_name,omitempty"`
	CreatorID    string     `json:"creator_id"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type StatusStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type AssigneeStat struct {
	AssigneeID   *string `json:"assignee_id,omitempty"`
	AssigneeName string  `json:"assignee_name"`
	Count        int     `json:"count"`
}
