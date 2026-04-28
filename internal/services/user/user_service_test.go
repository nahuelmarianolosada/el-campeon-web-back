package user

import (
	"errors"
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	users map[uint]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uint]*models.User),
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if user.Email == "duplicate@example.com" {
		return errors.New("email already registered")
	}
	user.ID = 1
	m.users[1] = user
	return nil
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) Update(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(id uint) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) FindAll(limit, offset int) ([]models.User, error) {
	var users []models.User
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

// Tests del UserService

func TestRegisterSuccess(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:     "test-secret",
		JWTRefreshSecret: "test-refresh",
		JWTExpiryHours:   24,
	}

	repo := NewMockUserRepository()
	service := NewUserService(repo, cfg)

	req := &models.RegisterRequest{
		Email:     "newuser@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "SecurePassword123!",
		Phone:     "+5491123456789",
	}

	resp, err := service.Register(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if resp.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	if resp.User.Email != "newuser@example.com" {
		t.Errorf("Expected email newuser@example.com, got %s", resp.User.Email)
	}

	if resp.User.Role != "USER" {
		t.Errorf("Expected role USER, got %s", resp.User.Role)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:     "test-secret",
		JWTRefreshSecret: "test-refresh",
		JWTExpiryHours:   24,
	}

	repo := NewMockUserRepository()
	service := NewUserService(repo, cfg)

	req := &models.RegisterRequest{
		Email:     "duplicate@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "SecurePassword123!",
	}

	_, err := service.Register(req)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
}

func TestLoginSuccess(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:     "test-secret",
		JWTRefreshSecret: "test-refresh",
		JWTExpiryHours:   24,
	}

	repo := NewMockUserRepository()
	service := NewUserService(repo, cfg)

	// Registrar usuario primero
	registerReq := &models.RegisterRequest{
		Email:     "login@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "SecurePassword123!",
	}

	service.Register(registerReq)

	// Intentar login
	loginReq := &models.LoginRequest{
		Email:    "login@example.com",
		Password: "SecurePassword123!",
	}

	resp, err := service.Login(loginReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:     "test-secret",
		JWTRefreshSecret: "test-refresh",
		JWTExpiryHours:   24,
	}

	repo := NewMockUserRepository()
	service := NewUserService(repo, cfg)

	loginReq := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "AnyPassword",
	}

	_, err := service.Login(loginReq)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}
