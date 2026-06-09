package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/guest"
)

type GuestHandler struct {
	guestService guest.GuestService
}

func NewGuestHandler(guestService guest.GuestService) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

// VerifyEmail envía un código de verificación al email
// @Summary Solicitar código de verificación
// @Tags Guest
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "Email a verificar"
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Failure 429 {object} gin.H
// @Router /api/guest/verify-email [post]
func (h *GuestHandler) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := c.ClientIP()
	err := h.guestService.VerifyEmailAndSendCode(req.Email, clientIP)
	if err != nil {
		// Revisar si es error de rate limit
		if err.Error() == "too many verification attempts from this IP. Try again in 15 minutes" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Verification code sent to your email",
		"expires_in_seconds": 600, // 10 minutos
	})
}

// ConfirmEmail confirms email and creates guest session
// @Summary Confirmar email con código
// @Tags Guest
// @Accept json
// @Produce json
// @Param request body models.ConfirmEmailRequest true "Email y código de verificación"
// @Success 200 {object} models.GuestSessionResponse
// @Failure 500 {object} gin.H
// @Router /api/guest/confirm-email [post]
func (h *GuestHandler) ConfirmEmail(c *gin.Context) {
	var req models.ConfirmEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientIP := c.ClientIP()
	sessionResp, err := h.guestService.ConfirmEmailAndCreateSession(req.Email, req.VerificationCode, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessionResp)
}
