package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"taskflow/backend/internal/store"
)

type createTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (h *Handler) ListTasks(c *gin.Context) {
	projectID := c.Param("id")
	userID := currentUserID(c)

	if _, err := h.store.GetProjectByID(c.Request.Context(), projectID); err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get project for tasks", "error", err)
		internalServerError(c)
		return
	}

	allowed, err := h.store.UserCanAccessProject(c.Request.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("check project access for tasks", "error", err)
		internalServerError(c)
		return
	}
	if !allowed {
		forbidden(c)
		return
	}

	statusFilter := strings.TrimSpace(c.Query("status"))
	if statusFilter != "" && !validStatus(statusFilter) {
		validationError(c, map[string]string{"status": "must be one of: todo, in_progress, done"})
		return
	}

	assigneeFilter := strings.TrimSpace(c.Query("assignee"))

	pagination, fieldErrs := parsePagination(c)
	if fieldErrs != nil {
		validationError(c, fieldErrs)
		return
	}

	tasks, err := h.store.ListTasksForProject(c.Request.Context(), projectID, statusFilter, assigneeFilter, pagination)
	if err != nil {
		h.logger.Error("list tasks", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *Handler) CreateTask(c *gin.Context) {
	projectID := c.Param("id")
	userID := currentUserID(c)

	if _, err := h.store.GetProjectByID(c.Request.Context(), projectID); err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get project for create task", "error", err)
		internalServerError(c)
		return
	}

	allowed, err := h.store.UserCanAccessProject(c.Request.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("check project access for create task", "error", err)
		internalServerError(c)
		return
	}
	if !allowed {
		forbidden(c)
		return
	}

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}

	fields := map[string]string{}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = normalizeOptionalString(req.Description)
	if req.AssigneeID != nil {
		trimmed := strings.TrimSpace(*req.AssigneeID)
		req.AssigneeID = &trimmed
		if trimmed == "" {
			req.AssigneeID = nil
		}
	}

	if req.Title == "" {
		fields["title"] = "is required"
	}

	if req.Status == "" {
		req.Status = "todo"
	}
	if !validStatus(req.Status) {
		fields["status"] = "must be one of: todo, in_progress, done"
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}
	if !validPriority(req.Priority) {
		fields["priority"] = "must be one of: low, medium, high"
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := parseDate(*req.DueDate)
		if err != nil {
			fields["due_date"] = "must be YYYY-MM-DD"
		} else {
			dueDate = parsed
		}
	}

	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	task, err := h.store.CreateTask(c.Request.Context(), store.CreateTaskParams{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		ProjectID:   projectID,
		AssigneeID:  req.AssigneeID,
		CreatorID:   userID,
		DueDate:     dueDate,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				validationError(c, map[string]string{"assignee_id": "does not reference a valid user"})
				return
			}
			if pgErr.Code == "22P02" {
				validationError(c, map[string]string{"assignee_id": "must be a valid UUID"})
				return
			}
		}
		h.logger.Error("create task", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := currentUserID(c)

	accessInfo, err := h.store.GetTaskAccessInfo(c.Request.Context(), taskID)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get task access info", "error", err)
		internalServerError(c)
		return
	}

	canUpdate := userID == accessInfo.ProjectOwner || userID == accessInfo.CreatorID
	if accessInfo.CurrentAssign != nil && *accessInfo.CurrentAssign == userID {
		canUpdate = true
	}
	if !canUpdate {
		forbidden(c)
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}
	if len(payload) == 0 {
		validationError(c, map[string]string{"body": "at least one field is required"})
		return
	}

	fields := map[string]string{}
	params := store.UpdateTaskParams{}

	for key, value := range payload {
		switch key {
		case "title":
			str, ok := value.(string)
			if !ok {
				fields["title"] = "must be a string"
				continue
			}
			trimmed := strings.TrimSpace(str)
			if trimmed == "" {
				fields["title"] = "cannot be empty"
				continue
			}
			params.Title = &trimmed
		case "description":
			params.Description.Set = true
			if value == nil {
				params.Description.Value = nil
				continue
			}
			str, ok := value.(string)
			if !ok {
				fields["description"] = "must be a string"
				continue
			}
			trimmed := strings.TrimSpace(str)
			if trimmed == "" {
				params.Description.Value = nil
			} else {
				params.Description.Value = &trimmed
			}
		case "status":
			str, ok := value.(string)
			if !ok || !validStatus(str) {
				fields["status"] = "must be one of: todo, in_progress, done"
				continue
			}
			params.Status = &str
		case "priority":
			str, ok := value.(string)
			if !ok || !validPriority(str) {
				fields["priority"] = "must be one of: low, medium, high"
				continue
			}
			params.Priority = &str
		case "assignee_id":
			params.AssigneeID.Set = true
			if value == nil {
				params.AssigneeID.Value = nil
				continue
			}
			str, ok := value.(string)
			if !ok {
				fields["assignee_id"] = "must be a string"
				continue
			}
			trimmed := strings.TrimSpace(str)
			if trimmed == "" {
				params.AssigneeID.Value = nil
			} else {
				params.AssigneeID.Value = &trimmed
			}
		case "due_date":
			params.DueDate.Set = true
			if value == nil {
				params.DueDate.Value = nil
				continue
			}
			str, ok := value.(string)
			if !ok {
				fields["due_date"] = "must be a string in YYYY-MM-DD format"
				continue
			}
			parsed, err := parseDate(str)
			if err != nil {
				fields["due_date"] = "must be YYYY-MM-DD"
				continue
			}
			params.DueDate.Value = parsed
		default:
			fields[key] = "is not supported"
		}
	}

	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	task, err := h.store.UpdateTask(c.Request.Context(), taskID, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				validationError(c, map[string]string{"assignee_id": "does not reference a valid user"})
				return
			}
			if pgErr.Code == "22P02" {
				validationError(c, map[string]string{"assignee_id": "must be a valid UUID"})
				return
			}
		}
		if err.Error() == "no fields to update" {
			validationError(c, map[string]string{"body": "at least one valid field is required"})
			return
		}
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("update task", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := currentUserID(c)

	accessInfo, err := h.store.GetTaskAccessInfo(c.Request.Context(), taskID)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get task for delete", "error", err)
		internalServerError(c)
		return
	}

	if userID != accessInfo.ProjectOwner && userID != accessInfo.CreatorID {
		forbidden(c)
		return
	}

	if err := h.store.DeleteTask(c.Request.Context(), taskID); err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("delete task", "error", err)
		internalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
