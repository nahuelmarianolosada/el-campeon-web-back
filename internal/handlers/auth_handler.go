package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/user"
)

type AuthHandler struct {
	userService user.UserService
}

func NewAuthHandler(userService user.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register registra un nuevo usuario con rol USER
// @Summary Registrar un nuevo usuario
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Datos de registro"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Conflict - Email already registered"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Register: Invalid request format - %v", err)
		invalidErr := ErrorMap[err]
		c.JSON(invalidErr.StatusCode, invalidErr)
		return
	}

	authResp, err := h.userService.Register(&req)
	if err != nil {
		// Detectar errores específicos de negocio
		errCode := "registration_failed"

		errResponse, errFound := ErrorMap[err]
		if !errFound {
			errResponse = ErrorResponse{
				Error:      errCode,
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}

		log.Printf("Register: %s - %v", errCode, err)
		c.JSON(errResponse.StatusCode, errResponse)
		return
	}

	log.Printf("Register: New user created successfully")
	c.JSON(http.StatusCreated, authResp)
}

// RegisterAdmin registra un nuevo usuario con rol personalizado (ADMIN/USER)
// Solo debe ser accesible desde un middleware de autorización ADMIN
// @Summary Registrar nuevo usuario (Admin)
// @Tags Auth
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.RegisterAdminRequest true "Datos de usuario a crear"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 409 {object} ErrorResponse "Conflict - Email already registered"
// @Router /auth/register-admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	var req models.RegisterAdminRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("RegisterAdmin: Invalid request format - %v", err)
		invalidErr := ErrorMap[err]
		c.JSON(invalidErr.StatusCode, invalidErr)
		return
	}

	authResp, err := h.userService.RegisterAdmin(&req)
	if err != nil {
		errCode := "user_creation_failed"

		errResponse, errFound := ErrorMap[err]
		if !errFound {
			errResponse = ErrorResponse{
				Error:      errCode,
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}

		log.Printf("RegisterAdmin: %s - %v", errCode, err)
		c.JSON(errResponse.StatusCode, errResponse)
		return
	}

	log.Printf("RegisterAdmin: New admin/user created successfully - Role: %s", req.Role)
	c.JSON(http.StatusCreated, authResp)
}

// Login autentica un usuario con credenciales
// @Summary Login de usuario
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Credenciales (email, password)"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login: Invalid request format - %v", err)
		errInvalidRequest := ErrorMap[err]
		c.JSON(errInvalidRequest.StatusCode, errInvalidRequest)
		return
	}

	authResp, err := h.userService.Login(&req)
	if err != nil {
		errAuthFailed := ErrorMap[err]

		log.Printf("Login: %s for email %s", errAuthFailed.Message, req.Email)
		c.JSON(errAuthFailed.StatusCode, errAuthFailed)
		return
	}

	log.Printf("Login: User authenticated successfully")
	c.JSON(http.StatusOK, authResp)
}

// RefreshToken renueva el access token usando el refresh token
// @Summary Renovar access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or expired token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {

	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("RefreshToken: Invalid request format - %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Missing or invalid refresh_token: " + err.Error(),
		})
		return
	}

	authResp, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("RefreshToken: Token refresh failed - %v", err)
		errInvalidRefreshToken := ErrorMap[err]
		c.JSON(errInvalidRefreshToken.StatusCode, errInvalidRefreshToken)
		return
	}

	log.Printf("RefreshToken: Tokens refreshed successfully")
	c.JSON(http.StatusOK, authResp)
}
