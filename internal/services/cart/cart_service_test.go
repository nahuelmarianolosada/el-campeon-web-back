package cart

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

// ============ Mock Repositories ============

// MockCartRepository
type MockCartRepository struct {
	carts         map[uint]*models.Cart
	items         map[uint]*models.CartItem
	nextID        uint
	nextItemID    uint
	GetCartErr    error
	AddItemErr    error
	UpdateItemErr error
	RemoveItemErr error
	ClearCartErr  error
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts:      make(map[uint]*models.Cart),
		items:      make(map[uint]*models.CartItem),
		nextID:     1,
		nextItemID: 1,
	}
}

func (m *MockCartRepository) GetOrCreateCart(userID uint) (*models.Cart, error) {
	if m.GetCartErr != nil {
		return nil, m.GetCartErr
	}

	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart, nil
		}
	}

	newCart := &models.Cart{
		ID:     m.nextID,
		UserID: userID,
		Items:  []models.CartItem{},
	}
	m.nextID++
	m.carts[newCart.ID] = newCart
	return newCart, nil
}

func (m *MockCartRepository) GetCart(userID uint) (*models.Cart, error) {
	if m.GetCartErr != nil {
		return nil, m.GetCartErr
	}

	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockCartRepository) AddItem(userID uint, item *models.CartItem) error {
	if m.AddItemErr != nil {
		return m.AddItemErr
	}

	item.ID = m.nextItemID
	m.nextItemID++
	m.items[item.ID] = item

	for _, cart := range m.carts {
		if cart.UserID == userID {
			cart.Items = append(cart.Items, *item)
			break
		}
	}
	return nil
}

func (m *MockCartRepository) UpdateItem(itemID uint, quantity int) error {
	if m.UpdateItemErr != nil {
		return m.UpdateItemErr
	}

	if item, exists := m.items[itemID]; exists {
		item.Quantity = quantity
	}
	return nil
}

func (m *MockCartRepository) RemoveItem(itemID uint) error {
	if m.RemoveItemErr != nil {
		return m.RemoveItemErr
	}

	delete(m.items, itemID)
	return nil
}

func (m *MockCartRepository) ClearCart(userID uint) error {
	if m.ClearCartErr != nil {
		return m.ClearCartErr
	}

	for _, cart := range m.carts {
		if cart.UserID == userID {
			cart.Items = []models.CartItem{}
			break
		}
	}
	return nil
}

func (m *MockCartRepository) GetCartItems(userID uint) ([]models.CartItem, error) {
	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart.Items, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// MockProductRepository
type MockProductRepository struct {
	products map[uint]*models.Product
	FindErr  error
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		products: make(map[uint]*models.Product),
	}
}

func (m *MockProductRepository) FindByID(id uint) (*models.Product, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}

	if product, exists := m.products[id]; exists {
		return product, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockProductRepository) FindBySKU(sku string) (*models.Product, error) {
	for _, product := range m.products {
		if product.SKU == sku {
			return product, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockProductRepository) Create(product *models.Product) error {
	return nil
}

func (m *MockProductRepository) Update(product *models.Product) error {
	return nil
}

func (m *MockProductRepository) Delete(id uint) error {
	delete(m.products, id)
	return nil
}

func (m *MockProductRepository) FindAll(limit, offset int) ([]models.Product, error) {
	return []models.Product{}, nil
}

func (m *MockProductRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	return []models.Product{}, nil
}

func (m *MockProductRepository) FindActive(limit, offset int) ([]models.Product, error) {
	return []models.Product{}, nil
}

func (m *MockProductRepository) UpdateStock(id uint, quantity int) error {
	return nil
}

// ============ Tests ============

func TestAddToCart_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup: Crear producto
	product := &models.Product{
		ID:              1,
		Name:            "Test Product",
		PriceRetail:     100.0,
		PriceWholesale:  80.0,
		Stock:           50,
		MinBulkQuantity: 10,
	}
	productRepo.products[1] = product

	// Agregar al carrito
	req := &models.AddToCartRequest{
		ProductID: 1,
		Quantity:  5,
	}

	err := service.AddToCart(1, req, false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verificar que se agregó
	cart, _ := cartRepo.GetCart(1)
	if len(cart.Items) != 1 {
		t.Fatalf("Expected 1 item in cart, got %d", len(cart.Items))
	}

	if cart.Items[0].Quantity != 5 {
		t.Fatalf("Expected quantity 5, got %d", cart.Items[0].Quantity)
	}

	if cart.Items[0].Price != 100.0 {
		t.Fatalf("Expected price 100.0 (retail), got %.2f", cart.Items[0].Price)
	}
}

func TestAddToCart_WithBulkPricing(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup: Crear producto
	product := &models.Product{
		ID:              1,
		Name:            "Test Product",
		PriceRetail:     100.0,
		PriceWholesale:  80.0,
		Stock:           50,
		MinBulkQuantity: 10,
	}
	productRepo.products[1] = product

	// Agregar al carrito como mayorista con cantidad >= MinBulkQuantity
	req := &models.AddToCartRequest{
		ProductID: 1,
		Quantity:  15,
	}

	err := service.AddToCart(1, req, true) // isBulkBuyer = true

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verificar que se aplicó precio mayorista
	cart, _ := cartRepo.GetCart(1)
	if cart.Items[0].Price != 80.0 {
		t.Fatalf("Expected wholesale price 80.0, got %.2f", cart.Items[0].Price)
	}
}

func TestAddToCart_InsufficientStock(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup: Producto con poco stock
	product := &models.Product{
		ID:    1,
		Stock: 5,
	}
	productRepo.products[1] = product

	// Intentar agregar más de lo disponible
	req := &models.AddToCartRequest{
		ProductID: 1,
		Quantity:  10,
	}

	err := service.AddToCart(1, req, false)

	if err == nil {
		t.Fatal("Expected error for insufficient stock")
	}
}

func TestAddToCart_ProductNotFound(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Intentar agregar producto que no existe
	req := &models.AddToCartRequest{
		ProductID: 999,
		Quantity:  5,
	}

	err := service.AddToCart(1, req, false)

	if err == nil {
		t.Fatal("Expected error for product not found")
	}
}

func TestGetCart_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup: Crear carrito con items
	cart, _ := cartRepo.GetOrCreateCart(1)
	item1 := &models.CartItem{
		CartID:    cart.ID,
		ProductID: 1,
		Quantity:  2,
		Price:     100.0,
		Product:   &models.Product{ID: 1, Name: "Product 1"},
	}
	item2 := &models.CartItem{
		CartID:    cart.ID,
		ProductID: 2,
		Quantity:  3,
		Price:     50.0,
		Product:   &models.Product{ID: 2, Name: "Product 2"},
	}
	errItem1 := cartRepo.AddItem(1, item1)
	if errItem1 != nil {
		t.Error("Error adding item1 to cart:", errItem1)
	}
	errItem2 := cartRepo.AddItem(1, item2)
	if errItem2 != nil {
		t.Error("Error adding item1 to cart:", errItem2)
	}

	// Obtener carrito
	resp, err := service.GetCart(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(resp.Items))
	}

	if resp.Total != 350.0 { // (2 * 100) + (3 * 50)
		t.Fatalf("Expected total 350.0, got %.2f", resp.Total)
	}
}

func TestUpdateCartItem_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup
	product := &models.Product{
		ID:    1,
		Stock: 100,
	}
	productRepo.products[1] = product

	cart, _ := cartRepo.GetOrCreateCart(1)
	item := &models.CartItem{
		CartID:    cart.ID,
		ProductID: 1,
		Quantity:  5,
		Price:     100.0,
	}
	errAddItem := cartRepo.AddItem(1, item)
	if errAddItem != nil {
		t.Error("Error adding item to cart:", errAddItem)
	}

	// Actualizar cantidad
	err := service.UpdateCartItem(1, item.ID, 10, false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cartRepo.items[item.ID].Quantity != 10 {
		t.Fatalf("Expected quantity 10, got %d", cartRepo.items[item.ID].Quantity)
	}
}

func TestRemoveFromCart_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup
	cart, _ := cartRepo.GetOrCreateCart(1)
	item := &models.CartItem{CartID: cart.ID, ProductID: 1, Quantity: 5}
	errAddItem := cartRepo.AddItem(1, item)
	if errAddItem != nil {
		t.Error("Error adding item to cart:", errAddItem)
	}

	// Remover
	err := service.RemoveFromCart(1, item.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, exists := cartRepo.items[item.ID]; exists {
		t.Fatal("Expected item to be removed")
	}
}

func TestClearCart_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup
	cart, _ := cartRepo.GetOrCreateCart(1)
	item := &models.CartItem{CartID: cart.ID, ProductID: 1, Quantity: 5}
	errAddItem := cartRepo.AddItem(1, item)
	if errAddItem != nil {
		t.Error("Error adding item to cart:", errAddItem)
	}

	// Limpiar
	err := service.ClearCart(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(cart.Items) != 0 {
		t.Fatalf("Expected empty cart, got %d items", len(cart.Items))
	}
}

func TestCalculateCartTotal_Success(t *testing.T) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup
	cart, _ := cartRepo.GetOrCreateCart(1)
	item1 := &models.CartItem{CartID: cart.ID, ProductID: 1, Quantity: 2, Price: 100.0}
	item2 := &models.CartItem{CartID: cart.ID, ProductID: 2, Quantity: 3, Price: 50.0}
	errAddItem1 := cartRepo.AddItem(1, item1)
	if errAddItem1 != nil {
		t.Error("Error adding item1 to cart:", errAddItem1)
	}
	errAddItem2 := cartRepo.AddItem(1, item2)
	if errAddItem2 != nil {
		t.Error("Error adding item2 to cart:", errAddItem2)
	}

	// Calcular total
	total, err := service.CalculateCartTotal(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if total != 350.0 {
		t.Fatalf("Expected 350.0, got %.2f", total)
	}
}

// ============ Benchmark Tests ============

func BenchmarkAddToCart(b *testing.B) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	product := &models.Product{
		ID:    1,
		Stock: 1000,
	}
	productRepo.products[1] = product

	req := &models.AddToCartRequest{
		ProductID: 1,
		Quantity:  5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.AddToCart(uint(i), req, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCart(b *testing.B) {
	cartRepo := NewMockCartRepository()
	productRepo := NewMockProductRepository()

	service := NewCartService(cartRepo, productRepo)

	// Setup: Crear carrito con items
	cart, _ := cartRepo.GetOrCreateCart(1)
	for i := 0; i < 10; i++ {
		item := &models.CartItem{
			CartID:    cart.ID,
			ProductID: uint(i + 1),
			Quantity:  5,
			Price:     100.0,
		}
		err := cartRepo.AddItem(1, item)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetCart(1)
		if err != nil {
			b.Fatal(err)
		}
	}
}
