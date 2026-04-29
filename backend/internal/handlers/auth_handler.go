package handlers

import (
	"net/http"

	"github.com/atakhanov/sensory-navigator/backend/internal/auth"
	"github.com/atakhanov/sensory-navigator/backend/internal/config"
	"github.com/atakhanov/sensory-navigator/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	users *services.UserService
	cfg   *config.Config
}

func NewAuthHandler(users *services.UserService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

type registerReq struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type tokenResp struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	User      any    `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Register(services.RegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	token, exp, err := auth.IssueToken(user.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tokenResp{
		Token:     token,
		ExpiresAt: exp.Format("2006-01-02T15:04:05Z07:00"),
		User:      user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, exp, err := auth.IssueToken(user.ID, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tokenResp{
		Token:     token,
		ExpiresAt: exp.Format("2006-01-02T15:04:05Z07:00"),
		User:      user,
	})
}
