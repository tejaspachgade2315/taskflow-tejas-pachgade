package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.store.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("list users", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}
