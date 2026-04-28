package user

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	errService "github.com/nahuelmarianolosada/el-campeon-web/internal/services/errors"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/utils"
	"gorm.io/gorm"
)

// Constantes de validación y roles
const (
	RoleUser          = "USER"
	RoleAdmin         = "ADMIN"
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxEmailLength    = 255
	MaxNameLength     = 100
)

type UserService interface {
	Register(req *models.RegisterRequest) (*models.AuthResponse, error)
	RegisterAdmin(req *models.RegisterAdminRequest) (*models.AuthResponse, error)
	Login(req *models.LoginRequest) (*models.AuthResponse, error)
	GetUserByID(id uint) (*models.UserResponse, error)
	GetUserByEmail(email string) (*models.UserResponse, error)
	UpdateUser(id uint, user *models.User) (*models.UserResponse, error)
	RefreshToken(refreshToken string) (*models.AuthResponse, error)
	ListUsers(limit, offset int) ([]models.UserResponse, error)
	SetBulkBuyer(userID uint, isBulk bool) error
}

type userService struct {
	userRepo repositories.UserRepository
	config   *config.Config
}

func NewUserService(userRepo repositories.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// Register registra un nuevo usuario con rol USER
// Valida email único, contraseña segura y datos completos
func (s *userService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Normalizar y validar inputs
	if err := s.validateRegisterRequest(req); err != nil {
		log.Printf("Register validation failed for email %s: %v", req.Email, err)
		return nil, err
	}

	// Crear usuario con rol USER
	user, err := s.createUserWithRole(req.Email, req.FirstName, req.LastName, req.Password, RoleUser, req)
	if err != nil {
		log.Printf("Failed to create user with email %s: %v", req.Email, err)
		return nil, err
	}

	// Generar respuesta con tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		log.Printf("Failed to generate auth response for user %d: %v", user.ID, err)
		return nil, err
	}

	log.Printf("User registered successfully: %s (ID: %d, Role: %s)", user.Email, user.ID, user.Role)
	return authResp, nil
}

// RegisterAdmin registra un nuevo usuario con rol ADMIN
// Solo debe ser llamado desde handlers protegidos por middleware admin
func (s *userService) RegisterAdmin(req *models.RegisterAdminRequest) (*models.AuthResponse, error) {
	// Validar request
	if err := s.validateRegisterRequest(&req.RegisterRequest); err != nil {
		log.Printf("RegisterAdmin validation failed for email %s: %v", req.Email, err)
		return nil, err
	}

	// Validar rol
	if err := s.validateRole(req.Role); err != nil {
		log.Printf("RegisterAdmin invalid role %s: %v", req.Role, err)
		return nil, err
	}

	// Crear usuario con rol solicitado
	user, err := s.createUserWithRole(req.Email, req.FirstName, req.LastName, req.Password, req.Role, &req.RegisterRequest)
	if err != nil {
		log.Printf("Failed to create admin user with email %s: %v", req.Email, err)
		return nil, err
	}

	// Generar respuesta con tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		log.Printf("Failed to generate auth response for admin user %d: %v", user.ID, err)
		return nil, err
	}

	log.Printf("Admin user registered successfully: %s (ID: %d, Role: %s)", user.Email, user.ID, user.Role)
	return authResp, nil
}

// createUserWithRole es una función privada que consolida la lógica de creación de usuario
// Evita duplicación de código entre Register y RegisterAdmin
func (s *userService) createUserWithRole(email, firstName, lastName, password, role string, req *models.RegisterRequest) (*models.User, error) {
	// Normalizar email
	email = strings.TrimSpace(strings.ToLower(email))

	// Verificar que el email sea único
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error checking email: %w", err)
	}
	if existingUser != nil {
		return nil, errService.ErrEmailExists
	}

	// Hash de la contraseña
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Crear entidad Usuario
	user := &models.User{
		Email:       email,
		FirstName:   strings.TrimSpace(firstName),
		LastName:    strings.TrimSpace(lastName),
		Password:    hashedPassword,
		Phone:       strings.TrimSpace(req.Phone),
		Address:     strings.TrimSpace(req.Address),
		City:        strings.TrimSpace(req.City),
		PostalCode:  strings.TrimSpace(req.PostalCode),
		Country:     strings.TrimSpace(req.Country),
		Role:        role,
		IsActive:    true,
		IsBulkBuyer: false,
	}

	// Persistir en BD
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("error creating user in database: %w", err)
	}

	return user, nil
}

// generateAuthResponse genera la respuesta de autenticación con tokens JWT
func (s *userService) generateAuthResponse(user *models.User) (*models.AuthResponse, error) {
	// Generar access token (corta duración)
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	// Generar refresh token (larga duración)
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.userToResponse(user),
		ExpiresIn:    int64(s.config.JWTExpiryHours * 3600),
	}, nil
}

// validateRegisterRequest valida los datos del request de registro
func (s *userService) validateRegisterRequest(req *models.RegisterRequest) error {
	// Validar email
	email := strings.TrimSpace(req.Email)
	if email == "" {
		return errService.ErrEmailInvalid
	}
	if len(email) > MaxEmailLength {
		return fmt.Errorf("email too long: max %d characters", MaxEmailLength)
	}

	// Validar nombre
	if strings.TrimSpace(req.FirstName) == "" {
		return errors.New("first name is required")
	}
	if len(req.FirstName) > MaxNameLength {
		return fmt.Errorf("first name too long: max %d characters", MaxNameLength)
	}

	// Validar apellido
	if strings.TrimSpace(req.LastName) == "" {
		return errors.New("last name is required")
	}
	if len(req.LastName) > MaxNameLength {
		return fmt.Errorf("last name too long: max %d characters", MaxNameLength)
	}

	// Validar contraseña
	if len(req.Password) < MinPasswordLength {
		return errService.ErrPasswordTooShort
	}
	if len(req.Password) > MaxPasswordLength {
		return fmt.Errorf("password too long: max %d characters", MaxPasswordLength)
	}

	return nil
}

// validateRole valida que el rol sea uno de los permitidos
func (s *userService) validateRole(role string) error {
	role = strings.TrimSpace(strings.ToUpper(role))
	if role != RoleUser && role != RoleAdmin {
		return errService.ErrInvalidRole
	}
	return nil
}

// Login autentica un usuario y genera tokens JWT
// Valida credenciales, estado del usuario y retorna tokens
func (s *userService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Normalizar email
	email := strings.TrimSpace(strings.ToLower(req.Email))

	// Buscar usuario por email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Login attempt with non-existent email: %s", email)
			return nil, errService.ErrInvalidCredentials
		}
		log.Printf("Database error during login for email %s: %v", email, err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verificar contraseña
	if !utils.VerifyPassword(req.Password, user.Password) {
		log.Printf("Failed password verification for user: %s", email)
		return nil, errService.ErrInvalidCredentials
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		log.Printf("Login attempt by inactive user: %s", email)
		return nil, errService.ErrUserInactive
	}

	// Generar respuesta con tokens
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		log.Printf("Failed to generate tokens for user %d: %v", user.ID, err)
		return nil, err
	}

	log.Printf("User logged in successfully: %s (ID: %d)", user.Email, user.ID)
	return authResp, nil
}

// userToResponse convierte una entidad User a UserResponse
// Esta función privada evita duplicación de mapeo
func (s *userService) userToResponse(user *models.User) models.UserResponse {
	return models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		Address:     user.Address,
		City:        user.City,
		PostalCode:  user.PostalCode,
		Country:     user.Country,
		Role:        user.Role,
		IsActive:    user.IsActive,
		IsBulkBuyer: user.IsBulkBuyer,
		CreatedAt:   user.CreatedAt,
	}
}

func (s *userService) GetUserByID(id uint) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	resp := s.userToResponse(user)
	return &resp, nil
}

func (s *userService) GetUserByEmail(email string) (*models.UserResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	resp := s.userToResponse(user)
	return &resp, nil
}

func (s *userService) UpdateUser(id uint, updates *models.User) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Actualizar solo los campos que no están vacíos (trim de espacios)
	if name := strings.TrimSpace(updates.FirstName); name != "" {
		user.FirstName = name
	}
	if name := strings.TrimSpace(updates.LastName); name != "" {
		user.LastName = name
	}
	if phone := strings.TrimSpace(updates.Phone); phone != "" {
		user.Phone = phone
	}
	if addr := strings.TrimSpace(updates.Address); addr != "" {
		user.Address = addr
	}
	if city := strings.TrimSpace(updates.City); city != "" {
		user.City = city
	}
	if postal := strings.TrimSpace(updates.PostalCode); postal != "" {
		user.PostalCode = postal
	}
	if country := strings.TrimSpace(updates.Country); country != "" {
		user.Country = country
	}

	if err = s.userRepo.Update(user); err != nil {
		log.Printf("Error updating user %d: %v", id, err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	log.Printf("User %d updated successfully", id)
	resp := s.userToResponse(user)
	return &resp, nil
}

func (s *userService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken, s.config)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Refresh token user not found: ID %d", claims.UserID)
			return nil, fmt.Errorf("user not found: %w", err)
		}
		log.Printf("Database error fetching user %d: %v", claims.UserID, err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !user.IsActive {
		log.Printf("Refresh token attempt by inactive user: %d", user.ID)
		return nil, errService.ErrUserInactive
	}

	// Generar nuevo access token (mantener mismo refresh token)
	authResp, err := s.generateAuthResponse(user)
	if err != nil {
		log.Printf("Failed to generate new tokens for user %d: %v", user.ID, err)
		return nil, err
	}

	log.Printf("Tokens refreshed for user %d", user.ID)
	return authResp, nil
}

func (s *userService) ListUsers(limit, offset int) ([]models.UserResponse, error) {
	users, err := s.userRepo.FindAll(limit, offset)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.userToResponse(&user)
	}

	return responses, nil
}

func (s *userService) SetBulkBuyer(userID uint, isBulk bool) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("database error: %w", err)
	}

	user.IsBulkBuyer = isBulk
	if err := s.userRepo.Update(user); err != nil {
		log.Printf("Error updating bulk buyer status for user %d: %v", userID, err)
		return fmt.Errorf("database error: %w", err)
	}

	action := "enabled"
	if !isBulk {
		action = "disabled"
	}
	log.Printf("Bulk buyer status %s for user %d", action, userID)
	return nil
}
