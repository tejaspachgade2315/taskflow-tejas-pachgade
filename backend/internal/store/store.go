package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

type OptionalString struct {
	Set   bool
	Value *string
}

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

type CreateTaskParams struct {
	Title       string
	Description *string
	Status      string
	Priority    string
	ProjectID   string
	AssigneeID  *string
	CreatorID   string
	DueDate     *time.Time
}

type UpdateTaskParams struct {
	Title       *string
	Description OptionalString
	Status      *string
	Priority    *string
	AssigneeID  OptionalString
	DueDate     OptionalTime
}

type TaskAccessInfo struct {
	TaskID        string
	ProjectID     string
	CreatorID     string
	ProjectOwner  string
	CurrentAssign *string
}

func (s *Store) CreateUser(ctx context.Context, name, email, passwordHash string) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`, name, strings.ToLower(email), passwordHash).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (UserWithPassword, error) {
	var user UserWithPassword
	err := s.db.QueryRow(ctx, `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = $1
	`, strings.ToLower(email)).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return UserWithPassword{}, err
	}
	return user, nil
}

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Store) UserCanAccessProject(ctx context.Context, projectID, userID string) (bool, error) {
	var allowed bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			WHERE p.id = $1
			AND (
				p.owner_id = $2
				OR EXISTS (
					SELECT 1
					FROM tasks t
					WHERE t.project_id = p.id
					AND (t.assignee_id = $2 OR t.creator_id = $2)
				)
			)
		)
	`, projectID, userID).Scan(&allowed)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func (s *Store) IsProjectOwner(ctx context.Context, projectID, userID string) (bool, error) {
	var owner bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM projects
			WHERE id = $1 AND owner_id = $2
		)
	`, projectID, userID).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner, nil
}

func (s *Store) CreateProject(ctx context.Context, name string, description *string, ownerID string) (Project, error) {
	var project Project
	err := s.db.QueryRow(ctx, `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, owner_id, created_at
	`, name, description, ownerID).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) ListProjectsForUser(ctx context.Context, userID string, p Pagination) ([]Project, error) {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.creator_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, p.Limit, p.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *Store) GetProjectByID(ctx context.Context, projectID string) (Project, error) {
	var project Project
	err := s.db.QueryRow(ctx, `
		SELECT id, name, description, owner_id, created_at
		FROM projects
		WHERE id = $1
	`, projectID).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) UpdateProject(ctx context.Context, projectID, ownerID string, name string, description *string) (Project, error) {
	var project Project
	err := s.db.QueryRow(ctx, `
		UPDATE projects
		SET name = $1, description = $2
		WHERE id = $3 AND owner_id = $4
		RETURNING id, name, description, owner_id, created_at
	`, name, description, projectID, ownerID).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) DeleteProject(ctx context.Context, projectID, ownerID string) error {
	cmd, err := s.db.Exec(ctx, `
		DELETE FROM projects
		WHERE id = $1 AND owner_id = $2
	`, projectID, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListTasksForProject(ctx context.Context, projectID string, statusFilter string, assigneeFilter string, p Pagination) ([]Task, error) {
	query := `
		SELECT
			t.id,
			t.title,
			t.description,
			t.status,
			t.priority,
			t.project_id,
			t.assignee_id,
			u.name,
			t.creator_id,
			t.due_date,
			t.created_at,
			t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.project_id = $1
	`

	args := []interface{}{projectID}
	placeholder := 2

	if statusFilter != "" {
		query += fmt.Sprintf(" AND t.status = $%d", placeholder)
		args = append(args, statusFilter)
		placeholder++
	}
	if assigneeFilter != "" {
		query += fmt.Sprintf(" AND t.assignee_id = $%d", placeholder)
		args = append(args, assigneeFilter)
		placeholder++
	}

	query += fmt.Sprintf(" ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d", placeholder, placeholder+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Store) CreateTask(ctx context.Context, params CreateTaskParams) (Task, error) {
	row := s.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, creator_id, due_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		)
		SELECT i.id, i.title, i.description, i.status, i.priority, i.project_id, i.assignee_id, u.name, i.creator_id, i.due_date, i.created_at, i.updated_at
		FROM inserted i
		LEFT JOIN users u ON u.id = i.assignee_id
	`, params.Title, params.Description, params.Status, params.Priority, params.ProjectID, params.AssigneeID, params.CreatorID, params.DueDate)

	task, err := scanTask(row)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (s *Store) GetTaskAccessInfo(ctx context.Context, taskID string) (TaskAccessInfo, error) {
	var info TaskAccessInfo
	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.project_id, t.creator_id, p.owner_id, t.assignee_id
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`, taskID).Scan(&info.TaskID, &info.ProjectID, &info.CreatorID, &info.ProjectOwner, &info.CurrentAssign)
	if err != nil {
		return TaskAccessInfo{}, err
	}
	return info, nil
}

func (s *Store) UpdateTask(ctx context.Context, taskID string, params UpdateTaskParams) (Task, error) {
	setParts := make([]string, 0)
	args := make([]interface{}, 0)
	placeholder := 1

	if params.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", placeholder))
		args = append(args, *params.Title)
		placeholder++
	}
	if params.Description.Set {
		if params.Description.Value == nil {
			setParts = append(setParts, "description = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("description = $%d", placeholder))
			args = append(args, *params.Description.Value)
			placeholder++
		}
	}
	if params.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", placeholder))
		args = append(args, *params.Status)
		placeholder++
	}
	if params.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", placeholder))
		args = append(args, *params.Priority)
		placeholder++
	}
	if params.AssigneeID.Set {
		if params.AssigneeID.Value == nil {
			setParts = append(setParts, "assignee_id = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("assignee_id = $%d", placeholder))
			args = append(args, *params.AssigneeID.Value)
			placeholder++
		}
	}
	if params.DueDate.Set {
		if params.DueDate.Value == nil {
			setParts = append(setParts, "due_date = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("due_date = $%d", placeholder))
			args = append(args, *params.DueDate.Value)
			placeholder++
		}
	}

	if len(setParts) == 0 {
		return Task{}, errors.New("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, taskID)

	query := fmt.Sprintf(`
		WITH updated AS (
			UPDATE tasks
			SET %s
			WHERE id = $%d
			RETURNING id, title, description, status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		)
		SELECT u.id, u.title, u.description, u.status, u.priority, u.project_id, u.assignee_id, usr.name, u.creator_id, u.due_date, u.created_at, u.updated_at
		FROM updated u
		LEFT JOIN users usr ON usr.id = u.assignee_id
	`, strings.Join(setParts, ", "), placeholder)

	row := s.db.QueryRow(ctx, query, args...)
	task, err := scanTask(row)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) error {
	cmd, err := s.db.Exec(ctx, `
		DELETE FROM tasks
		WHERE id = $1
	`, taskID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetProjectStats(ctx context.Context, projectID string) ([]StatusStat, []AssigneeStat, error) {
	statusRows, err := s.db.Query(ctx, `
		SELECT status, COUNT(*)::int
		FROM tasks
		WHERE project_id = $1
		GROUP BY status
		ORDER BY status
	`, projectID)
	if err != nil {
		return nil, nil, err
	}
	defer statusRows.Close()

	statusStats := make([]StatusStat, 0)
	for statusRows.Next() {
		var stat StatusStat
		if err := statusRows.Scan(&stat.Status, &stat.Count); err != nil {
			return nil, nil, err
		}
		statusStats = append(statusStats, stat)
	}
	if err := statusRows.Err(); err != nil {
		return nil, nil, err
	}

	assigneeRows, err := s.db.Query(ctx, `
		SELECT
			t.assignee_id,
			COALESCE(u.name, 'Unassigned') AS assignee_name,
			COUNT(*)::int
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.project_id = $1
		GROUP BY t.assignee_id, u.name
		ORDER BY assignee_name
	`, projectID)
	if err != nil {
		return nil, nil, err
	}
	defer assigneeRows.Close()

	assigneeStats := make([]AssigneeStat, 0)
	for assigneeRows.Next() {
		var stat AssigneeStat
		if err := assigneeRows.Scan(&stat.AssigneeID, &stat.AssigneeName, &stat.Count); err != nil {
			return nil, nil, err
		}
		assigneeStats = append(assigneeStats, stat)
	}
	if err := assigneeRows.Err(); err != nil {
		return nil, nil, err
	}

	return statusStats, assigneeStats, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanTask(s scanner) (Task, error) {
	var task Task
	var description sql.NullString
	var assigneeID sql.NullString
	var assigneeName sql.NullString
	var dueDate sql.NullTime

	err := s.Scan(
		&task.ID,
		&task.Title,
		&description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&assigneeID,
		&assigneeName,
		&task.CreatorID,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}

	if description.Valid {
		task.Description = &description.String
	}
	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}
	if assigneeName.Valid {
		task.AssigneeName = &assigneeName.String
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	return task, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
