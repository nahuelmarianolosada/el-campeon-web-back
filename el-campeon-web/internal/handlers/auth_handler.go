package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services"
)

type AuthHandler struct {
	userService services.UserService
}

func NewAuthHandler(userService services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register maneja el registro de usuarios
// @Summary Registrar un nuevo usuario
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Datos de registro"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} gin.H
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResp, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, authResp)
}

// Login maneja el login de usuarios
// @Summary Login de usuario
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Credenciales"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} gin.H
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResp, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// RefreshToken maneja la renovación de tokens
// @Summary Renovar access token usando refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <refresh_token>"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} gin.H
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResp, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

