package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"taskflow/backend/internal/auth"
	"taskflow/backend/internal/store"
)

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  store.User `json:"user"`
}

func (h *Handler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}

	fields := map[string]string{}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		fields["name"] = "is required"
	}
	if req.Email == "" {
		fields["email"] = "is required"
	} else if !strings.Contains(req.Email, "@") {
		fields["email"] = "must be a valid email"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	} else if len(req.Password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}

	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	hash, err := auth.HashPassword(req.Password, h.bcryptCost)
	if err != nil {
		h.logger.Error("hash password", "error", err)
		internalServerError(c)
		return
	}

	user, err := h.store.CreateUser(c.Request.Context(), req.Name, req.Email, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			validationError(c, map[string]string{"email": "already exists"})
			return
		}
		h.logger.Error("create user", "error", err)
		internalServerError(c)
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, h.jwtExpiry, user.ID, user.Email)
	if err != nil {
		h.logger.Error("generate token", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *Handler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}

	fields := map[string]string{}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		validationError(c, fields)
		return
	}

	user, err := h.store.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if store.IsNotFound(err) {
			unauthorized(c)
			return
		}
		h.logger.Error("get user by email", "error", err)
		internalServerError(c)
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		unauthorized(c)
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, h.jwtExpiry, user.ID, user.Email)
	if err != nil {
		h.logger.Error("generate token", "error", err)
		internalServerError(c)
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token: token,
		User: store.User{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}
