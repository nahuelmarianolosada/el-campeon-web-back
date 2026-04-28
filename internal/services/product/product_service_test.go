package product

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

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
	service := NewProductService(repo)

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

	service := NewProductService(repo)

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

	service := NewProductService(repo)

	// Usuario mayorista con cantidad >= mínimo
	price, err := service.GetPrice(1, true, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if price != 80.00 {
		t.Errorf("Expected price 80.00, got %.2f", price)
	}
}
