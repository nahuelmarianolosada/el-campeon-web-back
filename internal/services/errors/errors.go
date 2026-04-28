package errors

import "errors"

// ErrInvalidRequest Errores de la request
var (
	ErrInvalidRequest = errors.New("invalid request format")
)

// Errores estándar de negocio
var (
	ErrEmailExists         = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("missing or invalid refresh_token")
	ErrUserInactive        = errors.New("user account is inactive or not found")
	ErrInvalidRole         = errors.New("invalid role: must be USER or ADMIN")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters")
	ErrEmailInvalid        = errors.New("invalid email format")
)
