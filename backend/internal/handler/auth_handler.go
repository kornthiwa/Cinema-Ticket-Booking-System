package handler

import (
	"net/http"

	"cinema-booking/internal/auth"
	"cinema-booking/internal/model"
	"github.com/gin-gonic/gin"
)

// LoginRequest for mock auth (or Firebase ID token in production).
type LoginRequest struct {
	UserID string `json:"user_id"` // from Google/Firebase
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// Login returns JWT and upserts user. In production, verify Firebase ID token and extract claims.
func (h *Handler) Login(c *gin.Context) {
	var body LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.UserID == "" {
		body.UserID = body.Email
	}
	if body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id or email required"})
		return
	}
	role := model.RoleUser
	u := &model.User{ID: body.UserID, Email: body.Email, Name: body.Name, Role: role}
	if err := h.Repo.UpsertUser(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.IssueToken(h.JWTSecret, body.UserID, body.Email, string(role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": body.UserID, "role": string(role)})
}
