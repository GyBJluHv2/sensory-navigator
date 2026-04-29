package handlers

import (
	"net/http"

	"github.com/GyBJluHv2/sensory-navigator/backend/internal/middleware"
	"github.com/GyBJluHv2/sensory-navigator/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	users *services.UserService
}

func NewUsersHandler(s *services.UserService) *UsersHandler {
	return &UsersHandler{users: s}
}

// Me — GET /api/me
func (h *UsersHandler) Me(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
		return
	}
	user, err := h.users.GetByID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type updateProfileReq struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	NoisePref   int    `json:"noise_pref"`
	LightPref   int    `json:"light_pref"`
	CrowdPref   int    `json:"crowd_pref"`
}

// UpdateMe — PUT /api/me
func (h *UsersHandler) UpdateMe(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.UpdateProfile(uid, services.UpdateProfileInput{
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		NoisePref:   req.NoisePref,
		LightPref:   req.LightPref,
		CrowdPref:   req.CrowdPref,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword — PUT /api/me/password
func (h *UsersHandler) ChangePassword(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.users.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}