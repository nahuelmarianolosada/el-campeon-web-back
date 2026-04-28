package services_test

import (
	"errors"
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services"
	"gorm.io/gorm"
)

// MockUserRepository es un mock para las pruebas
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
	service := services.NewUserService(repo, cfg)

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
	service := services.NewUserService(repo, cfg)

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
	service := services.NewUserService(repo, cfg)

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
	service := services.NewUserService(repo, cfg)

	loginReq := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "AnyPassword",
	}

	_, err := service.Login(loginReq)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

// Mock Product Repository
type MockProductRepository struct {
	products map[uint]*models.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: make(map[uint]*models.Product),
	}
}

func (m *MockProductRepository) Create(product *models.Product) error {
	product.ID = 1
	m.products[1] = product
	return nil
}

func (m *MockProductRepository) FindByID(id uint) (*models.Product, error) {
	product, exists := m.products[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return product, nil
}

func (m *MockProductRepository) FindBySKU(sku string) (*models.Product, error) {
	for _, product := range m.products {
		if product.SKU == sku {
			return product, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockProductRepository) Update(product *models.Product) error {
	m.products[product.ID] = product
	return nil
}

func (m *MockProductRepository) Delete(id uint) error {
	delete(m.products, id)
	return nil
}

func (m *MockProductRepository) FindAll(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	for _, product := range m.products {
		products = append(products, *product)
	}
	return products, nil
}

func (m *MockProductRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	for _, product := range m.products {
		if product.Category == category {
			products = append(products, *product)
		}
	}
	return products, nil
}

func (m *MockProductRepository) FindActive(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	for _, product := range m.products {
		if product.IsActive {
			products = append(products, *product)
		}
	}
	return products, nil
}

func (m *MockProductRepository) UpdateStock(id uint, quantity int) error {
	product, exists := m.products[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	product.Stock += quantity
	return nil
}

// Tests del ProductService

func TestCreateProductSuccess(t *testing.T) {
	repo := NewMockProductRepository()
	service := services.NewProductService(repo)

	req := &models.CreateProductRequest{
		SKU:             "LIB-001",
		Name:            "Test Book",
		Description:     "A test book",
		Category:        "Books",
		PriceRetail:     100.00,
		PriceWholesale:  80.00,
		Stock:           50,
		MinBulkQuantity: 5,
	}

	resp, err := service.CreateProduct(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Name != "Test Book" {
		t.Errorf("Expected name Test Book, got %s", resp.Name)
	}

	if resp.PriceRetail != 100.00 {
		t.Errorf("Expected price 100.00, got %.2f", resp.PriceRetail)
	}
}

func TestGetPriceRetail(t *testing.T) {
	product := &models.Product{
		ID:              1,
		PriceRetail:     100.00,
		PriceWholesale:  80.00,
		MinBulkQuantity: 5,
	}

	repo := NewMockProductRepository()
	repo.products[1] = product

	service := services.NewProductService(repo)

	// Usuario regular con cantidad < mínimo
	price, err := service.GetPrice(1, false, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if price != 100.00 {
		t.Errorf("Expected price 100.00, got %.2f", price)
	}
}

func TestGetPriceWholesale(t *testing.T) {
	product := &models.Product{
		ID:              1,
		PriceRetail:     100.00,
		PriceWholesale:  80.00,
		MinBulkQuantity: 5,
	}

	repo := NewMockProductRepository()
	repo.products[1] = product

	service := services.NewProductService(repo)

	// Usuario mayorista con cantidad >= mínimo
	price, err := service.GetPrice(1, true, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if price != 80.00 {
		t.Errorf("Expected price 80.00, got %.2f", price)
	}
}

// Ejecutar tests
// go test ./internal/services/...

