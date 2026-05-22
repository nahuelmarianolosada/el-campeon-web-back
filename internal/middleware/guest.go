package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/guest"
)

// GuestAuthMiddleware valida token de sesión guest
func GuestAuthMiddleware(guestService guest.GuestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		guestToken := c.GetHeader("X-Guest-Token")
		if guestToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-Guest-Token header"})
			c.Abort()
			return
		}

		clientIP := c.ClientIP()
		session, err := guestService.ValidateGuestToken(guestToken, clientIP)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired guest token"})
			c.Abort()
			return
		}

		// Almacenar en contexto
		c.Set("guest_session_id", session.ID)
		c.Set("guest_email", session.Email)
		c.Set("user_id", session.UserID)
		c.Next()
	}
}

// OptionalAuthWithGuestMiddleware acepta tanto Authorization (Bearer) como X-Guest-Token
func OptionalAuthWithGuestMiddleware(cfg interface{}, guestService guest.GuestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Intentar guest token primero
		guestToken := c.GetHeader("X-Guest-Token")
		if guestToken != "" {
			clientIP := c.ClientIP()
			session, err := guestService.ValidateGuestToken(guestToken, clientIP)
			if err == nil {
				c.Set("guest_session_id", session.ID)
				c.Set("guest_email", session.Email)
				c.Set("user_id", session.UserID)
				c.Set("is_guest", true)
				c.Next()
				return
			}
		}

		// Fallback a Bearer token (auth normal)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// Aquí iría la validación de bearer token
		// Por ahora solo continuamos
		c.Next()
	}
}

