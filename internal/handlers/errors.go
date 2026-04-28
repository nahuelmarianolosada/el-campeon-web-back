package handlers

import (
	"net/http"

	errService "github.com/nahuelmarianolosada/el-campeon-web/internal/services/errors"
)

// ErrorResponse estructura estándar para errores
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"-"`
}

var ErrorMap = map[error]ErrorResponse{
	errService.ErrEmailExists: {
		Error:      "email_exists",
		Message:    errService.ErrEmailExists.Error(),
		StatusCode: http.StatusConflict,
	},
	errService.ErrInvalidRequest: {
		Error:      "invalid_request",
		Message:    errService.ErrInvalidRequest.Error(),
		StatusCode: http.StatusBadRequest,
	},
	errService.ErrInvalidRole: {
		Error:      "invalid_role",
		Message:    errService.ErrInvalidRole.Error(),
		StatusCode: http.StatusBadRequest,
	},
	errService.ErrInvalidCredentials: {
		Error:      "authentication_failed",
		Message:    errService.ErrInvalidCredentials.Error(),
		StatusCode: http.StatusUnauthorized,
	},
	errService.ErrInvalidRefreshToken: {
		Error:      "invalid_refresh_token",
		Message:    errService.ErrInvalidRefreshToken.Error(),
		StatusCode: http.StatusUnauthorized,
	},
}
