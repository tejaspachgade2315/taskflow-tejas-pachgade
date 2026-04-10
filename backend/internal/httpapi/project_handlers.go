package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/store"
)

type projectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *Handler) ListProjects(c *gin.Context) {
	pagination, fieldErrs := parsePagination(c)
	if fieldErrs != nil {
		validationError(c, fieldErrs)
		return
	}

	projects, err := h.store.ListProjectsForUser(c.Request.Context(), currentUserID(c), pagination)
	if err != nil {
		h.logger.Error("list projects", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = normalizeOptionalString(req.Description)

	fields := map[string]string{}
	if req.Name == "" {
		fields["name"] = "is required"
	}
	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	project, err := h.store.CreateProject(c.Request.Context(), req.Name, req.Description, currentUserID(c))
	if err != nil {
		h.logger.Error("create project", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *Handler) GetProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := currentUserID(c)

	project, err := h.store.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get project", "error", err)
		internalServerError(c)
		return
	}

	allowed, err := h.store.UserCanAccessProject(c.Request.Context(), projectID, userID)
	if err != nil {
		h.logger.Error("check project access", "error", err)
		internalServerError(c)
		return
	}
	if !allowed {
		forbidden(c)
		return
	}

	tasks, err := h.store.ListTasksForProject(c.Request.Context(), projectID, "", "", store.Pagination{Page: 1, Limit: 1000, Offset: 0})
	if err != nil {
		h.logger.Error("list project tasks", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"tasks":       tasks,
	})
}

func (h *Handler) UpdateProject(c *gin.Context) {
	projectID := c.Param("id")
	ownerID := currentUserID(c)

	isOwner, err := h.store.IsProjectOwner(c.Request.Context(), projectID, ownerID)
	if err != nil {
		h.logger.Error("check owner", "error", err)
		internalServerError(c)
		return
	}
	if !isOwner {
		project, err := h.store.GetProjectByID(c.Request.Context(), projectID)
		if err != nil {
			if store.IsNotFound(err) {
				notFound(c)
				return
			}
			h.logger.Error("get project for update", "error", err)
			internalServerError(c)
			return
		}
		if project.OwnerID != ownerID {
			forbidden(c)
			return
		}
	}

	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = normalizeOptionalString(req.Description)

	fields := map[string]string{}
	if req.Name == "" {
		fields["name"] = "is required"
	}
	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	project, err := h.store.UpdateProject(c.Request.Context(), projectID, ownerID, req.Name, req.Description)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("update project", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	projectID := c.Param("id")
	userID := currentUserID(c)

	project, err := h.store.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get project for delete", "error", err)
		internalServerError(c)
		return
	}

	if project.OwnerID != userID {
		forbidden(c)
		return
	}

	if err := h.store.DeleteProject(c.Request.Context(), projectID, userID); err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("delete project", "error", err)
		internalServerError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ProjectStats(c *gin.Context) {
	projectID := c.Param("id")
	userID := currentUserID(c)

	project, err := h.store.GetProjectByID(c.Request.Context(), projectID)
	if err != nil {
		if store.IsNotFound(err) {
			notFound(c)
			return
		}
		h.logger.Error("get project for stats", "error", err)
		internalServerError(c)
		return
	}

	allowed, err := h.store.UserCanAccessProject(c.Request.Context(), project.ID, userID)
	if err != nil {
		h.logger.Error("check access for stats", "error", err)
		internalServerError(c)
		return
	}
	if !allowed {
		forbidden(c)
		return
	}

	statusStats, assigneeStats, err := h.store.GetProjectStats(c.Request.Context(), project.ID)
	if err != nil {
		h.logger.Error("project stats", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id":   project.ID,
		"by_status":    statusStats,
		"by_assignee":  assigneeStats,
		"generated_at": time.Now().UTC(),
	})
}
