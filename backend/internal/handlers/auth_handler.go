package handlers

import (
	"net/http"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/auth"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/config"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	users         *services.UserService
	verifications *services.VerificationService
	cfg           *config.Config
}

func NewAuthHandler(users *services.UserService, verifications *services.VerificationService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, verifications: verifications, cfg: cfg}
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

// Register — POST /api/auth/register.
//
// Старый одношаговый эндпоинт регистрации, оставлен для обратной совместимости
// (cmd/seed, интеграционные тесты). Новый flow для UI — /register-request +
// /register-confirm с подтверждением кода из письма.
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

// RequestRegister — POST /api/auth/register-request.
//
// Шаг 1 регистрации с подтверждением email: проверка формата адреса,
// генерация 6-значного кода, отправка письма (или запись в лог
// в DEV-режиме), сохранение хэшей в таблице verification_codes.
func (h *AuthHandler) RequestRegister(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exp, err := h.verifications.Request(services.RequestRegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     "code_sent",
		"expires_at": exp.Format("2006-01-02T15:04:05Z07:00"),
	})
}

type confirmReq struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code"  binding:"required,len=6"`
}

// ConfirmRegister — POST /api/auth/register-confirm.
//
// Шаг 2 регистрации: пользователь вводит 6-значный код, сервер проверяет
// его, создаёт учётную запись с EmailVerified=true и выдаёт JWT.
func (h *AuthHandler) ConfirmRegister(c *gin.Context) {
	var req confirmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.verifications.Confirm(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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

type resendReq struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendCode — POST /api/auth/resend-code.
//
// Перевыпускает код для email, по которому уже был запущен RequestRegister.
// Прежний код инвалидируется, новый отправляется через тот же канал.
func (h *AuthHandler) ResendCode(c *gin.Context) {
	var req resendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exp, err := h.verifications.Resend(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":     "code_sent",
		"expires_at": exp.Format("2006-01-02T15:04:05Z07:00"),
	})
}