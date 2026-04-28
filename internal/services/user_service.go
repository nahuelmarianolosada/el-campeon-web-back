package services

import (
	"errors"
	"fmt"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/utils"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *models.RegisterRequest) (*models.AuthResponse, error)
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

func (s *userService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Verificar si el usuario ya existe
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash de la contraseña
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Crear nuevo usuario
	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  hashedPassword,
		Phone:     req.Phone,
		Address:   req.Address,
		City:      req.City,
		PostalCode: req.PostalCode,
		Country:   req.Country,
		Role:      "USER",
		IsActive:  true,
	}

	if err = s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Generar tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserResponse{
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
		},
		ExpiresIn: int64(s.config.JWTExpiryHours * 3600),
	}, nil
}

func (s *userService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	// Verificar contraseña
	if !utils.VerifyPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Verificar que el usuario está activo
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generar tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserResponse{
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
		},
		ExpiresIn: int64(s.config.JWTExpiryHours * 3600),
	}, nil
}

func (s *userService) GetUserByID(id uint) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	return &models.UserResponse{
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
	}, nil
}

func (s *userService) GetUserByEmail(email string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	return &models.UserResponse{
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
	}, nil
}

func (s *userService) UpdateUser(id uint, updates *models.User) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	// Actualizar solo los campos que no están vacíos
	if updates.FirstName != "" {
		user.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		user.LastName = updates.LastName
	}
	if updates.Phone != "" {
		user.Phone = updates.Phone
	}
	if updates.Address != "" {
		user.Address = updates.Address
	}
	if updates.City != "" {
		user.City = updates.City
	}
	if updates.PostalCode != "" {
		user.PostalCode = updates.PostalCode
	}
	if updates.Country != "" {
		user.Country = updates.Country
	}

	if err = s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &models.UserResponse{
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
	}, nil
}

func (s *userService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken, s.config)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generar nuevo access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserResponse{
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
		},
		ExpiresIn: int64(s.config.JWTExpiryHours * 3600),
	}, nil
}

func (s *userService) ListUsers(limit, offset int) ([]models.UserResponse, error) {
	users, err := s.userRepo.FindAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, models.UserResponse{
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
		})
	}

	return responses, nil
}

func (s *userService) SetBulkBuyer(userID uint, isBulk bool) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("error finding user: %w", err)
	}

	user.IsBulkBuyer = isBulk
	return s.userRepo.Update(user)
}

