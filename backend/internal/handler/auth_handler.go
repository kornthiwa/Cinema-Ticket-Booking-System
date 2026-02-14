package handler

import (
	"net/http"
	"strings"

	"cinema-booking/internal/auth"
	"cinema-booking/internal/model"
	"github.com/gin-gonic/gin"
)

// LoginRequest for email+password login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login verifies email+password and returns JWT.
func (h *Handler) Login(c *gin.Context) {
	var body LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	u, err := h.Repo.GetUserByEmail(c.Request.Context(), body.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if u.PasswordHash == "" || !auth.CheckPassword(u.PasswordHash, body.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	token, err := auth.IssueToken(h.JWTSecret, u.ID, u.Email, string(u.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": u.ID, "role": string(u.Role)})
}

// RegisterRequest for new user signup.
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Register creates a new user and returns JWT (auto-login).
func (h *Handler) Register(c *gin.Context) {
	var body RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	if len(body.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	_, err := h.Repo.GetUserByEmail(c.Request.Context(), body.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = body.Email
	}
	u := &model.User{
		ID:           body.Email,
		Email:        body.Email,
		Name:         name,
		Role:         model.RoleUser,
		PasswordHash: hash,
	}
	if err := h.Repo.UpsertUser(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.IssueToken(h.JWTSecret, u.ID, u.Email, string(u.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user_id": u.ID, "role": string(u.Role)})
}
