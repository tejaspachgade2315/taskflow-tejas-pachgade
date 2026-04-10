package httpapi

import "github.com/gin-gonic/gin"

func jsonError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func validationError(c *gin.Context, fields map[string]string) {
	c.JSON(400, gin.H{
		"error":  "validation failed",
		"fields": fields,
	})
}

func unauthorized(c *gin.Context) {
	jsonError(c, 401, "unauthorized")
}

func forbidden(c *gin.Context) {
	jsonError(c, 403, "forbidden")
}

func notFound(c *gin.Context) {
	jsonError(c, 404, "not found")
}

func internalServerError(c *gin.Context) {
	jsonError(c, 500, "internal server error")
}
