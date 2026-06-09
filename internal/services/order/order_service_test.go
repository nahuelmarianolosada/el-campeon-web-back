package order

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/order/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper para punteros a uint
func uintPtr(u uint) *uint {
	return &u
}

// ============ Mock Repositories ============

// MockOrderRepository
type MockOrderRepository struct {
	orders          map[uint]*models.Order
	nextID          uint
	CreateErr       error
	FindByIDErr     error
	FindByUserIDErr error
	FindAllErr      error
	UpdateStatusErr error
	AddItemErr      error
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[uint]*models.Order),
		nextID: 1,
	}
}

func (m *MockOrderRepository) Create(order *models.Order) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}

	order.ID = m.nextID
	m.nextID++
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) FindByID(id uint) (*models.Order, error) {
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}

	if order, exists := m.orders[id]; exists {
		return order, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockOrderRepository) FindByOrderNumber(orderNumber string) (*models.Order, error) {
	for _, order := range m.orders {
		if order.OrderNumber == orderNumber {
			return order, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockOrderRepository) FindByUserID(userID uint, _, _ int) ([]models.Order, error) {
	if m.FindByUserIDErr != nil {
		return nil, m.FindByUserIDErr
	}

	var orders []models.Order
	for _, order := range m.orders {
		if order.UserID != nil && *order.UserID == userID {
			orders = append(orders, *order)
		}
	}
	return orders, nil
}

func (m *MockOrderRepository) FindAll(_, _ int) ([]models.Order, error) {
	if m.FindAllErr != nil {
		return nil, m.FindAllErr
	}

	var orders []models.Order
	for _, order := range m.orders {
		orders = append(orders, *order)
	}
	return orders, nil
}

func (m *MockOrderRepository) UpdateStatus(orderID uint, status string) error {
	if m.UpdateStatusErr != nil {
		return m.UpdateStatusErr
	}

	if order, exists := m.orders[orderID]; exists {
		order.Status = status
	}
	return nil
}

func (m *MockOrderRepository) AddItem(orderID uint, item *models.OrderItem) error {
	if m.AddItemErr != nil {
		return m.AddItemErr
	}

	if order, exists := m.orders[orderID]; exists {
		order.Items = append(order.Items, *item)
	}
	return nil
}

func (m *MockOrderRepository) Update(order *models.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) Delete(id uint) error {
	delete(m.orders, id)
	return nil
}

// MockCartRepository para Order Tests
type MockOrderCartRepository struct {
	carts        map[uint]*models.Cart
	ClearCartErr error
}

func NewMockOrderCartRepository() *MockOrderCartRepository {
	return &MockOrderCartRepository{
		carts: make(map[uint]*models.Cart),
	}
}

func (m *MockOrderCartRepository) GetCart(userID uint) (*models.Cart, error) {
	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockOrderCartRepository) GetOrCreateCart(userID uint) (*models.Cart, error) {
	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockOrderCartRepository) AddItem(_ uint, _ *models.CartItem) error {
	return nil
}

func (m *MockOrderCartRepository) UpdateItem(_ uint, _ int) error {
	return nil
}

func (m *MockOrderCartRepository) UpdateCompleteItemInTheCart(_ *models.CartItem) error {
	return nil
}

func (m *MockOrderCartRepository) RemoveItem(_ uint) error {
	return nil
}

func (m *MockOrderCartRepository) ClearCart(userID uint) error {
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

func (m *MockOrderCartRepository) GetCartItems(userID uint) ([]models.CartItem, error) {
	for _, cart := range m.carts {
		if cart.UserID == userID {
			return cart.Items, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// MockUserRepository para Order Tests
type MockOrderUserRepository struct {
}

func (m *MockOrderUserRepository) FindByID(id uint) (*models.User, error) {
	return &models.User{ID: id}, nil
}

func (m *MockOrderUserRepository) FindByEmail(_ string) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *MockOrderUserRepository) Create(_ *models.User) error {
	return nil
}

func (m *MockOrderUserRepository) Update(_ *models.User) error {
	return nil
}

func (m *MockOrderUserRepository) Delete(_ uint) error {
	return nil
}

func (m *MockOrderUserRepository) FindAll(_, _ int) ([]models.User, error) {
	return []models.User{}, nil
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) FindByMercadopagoPaymentID(mercadopagoPaymentID string) (*models.Payment, error) {
	args := m.Called(mercadopagoPaymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByOrderNumber(orderNumber string) (*models.Payment, error) {
	args := m.Called(orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Create(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(id uint) (*models.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByOrderID(orderID uint) (*models.Payment, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByUserID(userID uint, limit, offset int) ([]models.Payment, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) ListAll(limit, offset int) ([]models.Payment, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(paymentID uint, status string) error {
	args := m.Called(paymentID, status)
	return args.Error(0)
}

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *models.Product) error { return nil }
func (m *MockProductRepository) FindByID(id uint) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}
func (m *MockProductRepository) FindBySKU(sku string) (*models.Product, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}
func (m *MockProductRepository) Update(product *models.Product) error { return nil }
func (m *MockProductRepository) Delete(id uint) error                 { return nil }
func (m *MockProductRepository) FindAll(limit, offset int) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) FindActive(limit, offset int) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) UpdateStock(id uint, quantity int) error { return nil }

type MockProductVariantRepository struct {
	mock.Mock
}

func (m *MockProductVariantRepository) CreateVariant(variant *models.ProductVariant) error {
	return nil
}
func (m *MockProductVariantRepository) FindVariantByID(id uint) (*models.ProductVariant, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) FindVariantsByProductID(productID uint) ([]models.ProductVariant, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) FindVariantsByProductIDAndValue(productID uint, value string) ([]models.ProductVariant, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) UpdateVariant(variant *models.ProductVariant) error {
	return nil
}
func (m *MockProductVariantRepository) DeleteVariant(id uint) error { return nil }
func (m *MockProductVariantRepository) CreateVariantValue(value *models.ProductVariantValue) (*models.ProductVariantValue, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) FindVariantValueByID(id uint) (*models.ProductVariantValue, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) FindVariantValuesByVariantID(variantID uint) ([]models.ProductVariantValue, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) UpdateVariantValue(value *models.ProductVariantValue) error {
	return nil
}
func (m *MockProductVariantRepository) DeleteVariantValue(id uint) error { return nil }
func (m *MockProductVariantRepository) DeleteVariantValuesByVariantID(variantID uint) error {
	return nil
}
func (m *MockProductVariantRepository) CreateVariantCombination(combination *models.ProductVariantCombination) error {
	return nil
}
func (m *MockProductVariantRepository) FindVariantCombinationByID(id uint) (*models.ProductVariantCombination, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantCombination), args.Error(1)
}
func (m *MockProductVariantRepository) FindVariantCombinationsByProductID(productID uint, limit int, offset int) ([]models.ProductVariantCombination, error) {
	return nil, nil
}
func (m *MockProductVariantRepository) FindVariantCombinationBySKU(sku string) (*models.ProductVariantCombination, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantCombination), args.Error(1)
}
func (m *MockProductVariantRepository) UpdateVariantCombination(combination *models.ProductVariantCombination) error {
	return nil
}
func (m *MockProductVariantRepository) DeleteVariantCombination(id uint) error { return nil }
func (m *MockProductVariantRepository) UpdateVariantCombinationStock(id uint, quantity int) error {
	return nil
}

type MockGuestRepository struct{}

func (m *MockGuestRepository) CreateGuestSession(session *models.GuestSession) error {
	return nil
}
func (m *MockGuestRepository) FindGuestSessionByEmail(email string) (*models.GuestSession, error) {
	return nil, nil
}
func (m *MockGuestRepository) UpdateGuestSession(session *models.GuestSession) error {
	return nil
}
func (m *MockGuestRepository) CountVerificationAttemptsByIP(ip string, minutesWindow int) (int, error) {
	return 0, nil
}
func (m *MockGuestRepository) DeleteExpiredSessions() (int64, error) {
	return 0, nil
}

type MockProductBranchStockRepository struct {
	mock.Mock
}

func (m *MockProductBranchStockRepository) FindByProduct(productID uint) ([]models.ProductBranchStock, error) {
	args := m.Called(productID)
	return args.Get(0).([]models.ProductBranchStock), args.Error(1)
}
func (m *MockProductBranchStockRepository) GetByProductAndBranch(productID, branchID uint) (*models.ProductBranchStock, error) {
	args := m.Called(productID, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductBranchStock), args.Error(1)
}
func (m *MockProductBranchStockRepository) Upsert(stock *models.ProductBranchStock) error {
	args := m.Called(stock)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) IncrementReserved(tx *gorm.DB, productID, branchID uint, delta int) error {
	args := m.Called(tx, productID, branchID, delta)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) DecrementStock(tx *gorm.DB, productID, branchID uint, qty int) error {
	args := m.Called(tx, productID, branchID, qty)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) BranchesWithFullStock(items []models.QuoteItem) ([]uint, error) {
	args := m.Called(items)
	return args.Get(0).([]uint), args.Error(1)
}
func (m *MockProductBranchStockRepository) OutOfStockItemsForBranch(branchID uint, items []models.QuoteItem) ([]uint, error) {
	args := m.Called(branchID, items)
	return args.Get(0).([]uint), args.Error(1)
}
func (m *MockProductBranchStockRepository) DB() *gorm.DB {
	return nil
}

// ============ Tests ============

func TestCreateOrder_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear carrito con items
	cart := &models.Cart{
		UserID: 1,
		Items: []models.CartItem{
			{
				ID:        1,
				ProductID: 1,
				Quantity:  2,
				Price:     100.0,
				Product:   &models.Product{ID: 1, Name: "Product 1"},
			},
			{
				ID:        2,
				ProductID: 2,
				Quantity:  3,
				Price:     50.0,
				Product:   &models.Product{ID: 2, Name: "Product 2"},
			},
		},
	}
	cartRepo.carts[1] = cart

	// Crear orden
	req := &models.CreateOrderRequest{
		ShippingAddress: map[string]interface{}{
			"street":      "123 Main St",
			"city":        "Buenos Aires",
			"postal_code": "1425",
			"country":     "Argentina",
		},
		Notes: "Test order",
	}

	resp, err := service.CreateOrder(1, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.UserID != 1 {
		t.Fatalf("Expected user_id 1, got %d", resp.UserID)
	}

	if resp.Subtotal != 350.0 { // (2 * 100) + (3 * 50)
		t.Fatalf("Expected subtotal 350.0, got %.2f", resp.Subtotal)
	}

	// Verificar que el carrito se limpió
	if len(cart.Items) != 0 {
		t.Fatalf("Expected empty cart after order creation, got %d items", len(cart.Items))
	}
}

func TestCreateOrder_EmptyCart(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Carrito vacío
	cart := &models.Cart{
		UserID: 1,
		Items:  []models.CartItem{},
	}
	cartRepo.carts[1] = cart

	req := &models.CreateOrderRequest{
		ShippingAddress: map[string]interface{}{
			"street": "123 Main St",
		},
	}

	_, err := service.CreateOrder(1, req)

	if err == nil {
		t.Fatal("Expected error for empty cart")
	}
}

func TestCreateOrder_CalculatesTax(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear carrito
	cart := &models.Cart{
		UserID: 1,
		Items: []models.CartItem{
			{
				ProductID: 1,
				Quantity:  1,
				Price:     100.0,
				Product:   &models.Product{ID: 1, Name: "Product 1"},
			},
		},
	}
	cartRepo.carts[1] = cart

	req := &models.CreateOrderRequest{
		ShippingAddress: map[string]interface{}{
			"street": "123 Main St",
		},
	}

	resp, err := service.CreateOrder(1, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Tax = 100 * 0.21 = 21
	if resp.Tax != 21.0 {
		t.Fatalf("Expected tax 21.0, got %.2f", resp.Tax)
	}

	// Total = 100 + 21 = 121
	if resp.Total != 121.0 {
		t.Fatalf("Expected total 121.0, got %.2f", resp.Total)
	}
}

func TestCreateGuestOrder_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup products
	product := &models.Product{
		ID:          1,
		Name:        "Simple Product",
		SKU:         "SIMPLE-123",
		PriceRetail: 100.0,
	}
	variantRepo.On("FindVariantCombinationBySKU", "SIMPLE-123").Return(nil, gorm.ErrRecordNotFound)
	productRepo.On("FindBySKU", "SIMPLE-123").Return(product, nil)

	req := &models.CreateGuestOrderRequest{
		GuestName:  "Guest User",
		GuestEmail: "guest@example.com",
		Items: []models.GuestCartItem{
			{
				SKU:      "SIMPLE-123",
				Quantity: 2,
				Price:    100.0,
			},
		},
		ShippingAddress: map[string]interface{}{
			"street": "Guest St 123",
		},
		DeliveryMethod: "shipping",
	}

	resp, err := service.CreateGuestOrder(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.GuestEmail != "guest@example.com" {
		t.Fatalf("Expected guest email guest@example.com, got %s", resp.GuestEmail)
	}

	if resp.Subtotal != 200.0 {
		t.Fatalf("Expected subtotal 200.0, got %.2f", resp.Subtotal)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}

	productRepo.AssertExpectations(t)
}

func TestGetOrderByID_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear orden
	order := &models.Order{
		OrderNumber: "ORD-12345",
		UserID:      uintPtr(1),
		Status:      status.Pending,
		Subtotal:    100.0,
		Tax:         21.0,
		Total:       121.0,
		Items: []models.OrderItem{
			{
				ProductID: 1,
				Quantity:  2,
				Price:     50.0,
				Product:   &models.Product{ID: 1, Name: "Product 1"},
			},
		},
	}
	err := orderRepo.Create(order)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Obtener orden
	resp, err := service.GetOrderByID(order.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.OrderNumber != "ORD-12345" {
		t.Fatalf("Expected order_number ORD-12345, got %s", resp.OrderNumber)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}
}

func TestGetOrdersByUserID_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear múltiples órdenes para el mismo usuario
	order1 := &models.Order{
		OrderNumber: "ORD-001",
		UserID:      uintPtr(1),
		Status:      status.Pending,
		Subtotal:    100.0,
		Tax:         21.0,
		Total:       121.0,
	}
	order2 := &models.Order{
		OrderNumber: "ORD-002",
		UserID:      uintPtr(1),
		Status:      status.Confirmed,
		Subtotal:    200.0,
		Tax:         42.0,
		Total:       242.0,
	}
	order3 := &models.Order{
		OrderNumber: "ORD-003",
		UserID:      uintPtr(2),
		Status:      status.Pending,
		Subtotal:    150.0,
		Tax:         31.5,
		Total:       181.5,
	}

	errOrder1 := orderRepo.Create(order1)
	errOrder2 := orderRepo.Create(order2)
	errOrder3 := orderRepo.Create(order3)
	if errOrder1 != nil {
		t.Fatalf("Expected no error, got %v", errOrder1)
	}
	if errOrder2 != nil {
		t.Fatalf("Expected no error, got %v", errOrder2)
	}
	if errOrder3 != nil {
		t.Fatalf("Expected no error, got %v", errOrder3)
	}

	// Obtener órdenes del usuario 1
	resp, err := service.GetOrdersByUserID(1, 10, 0)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 orders for user 1, got %d", len(resp))
	}
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear orden
	order := &models.Order{
		OrderNumber: "ORD-001",
		UserID:      uintPtr(1),
		Status:      status.Pending,
		Subtotal:    100.0,
		Tax:         21.0,
		Total:       121.0,
	}
	err := orderRepo.Create(order)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Mock the payment repo call
	paymentRepo.On("FindByOrderID", order.ID).Return(nil, nil)

	// Actualizar estado
	resp, err := service.UpdateOrderStatus(order.ID, status.Confirmed)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Status != status.Confirmed {
		t.Fatalf("Expected status CONFIRMED, got %s", resp.Status)
	}
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear orden
	order := &models.Order{
		OrderNumber: "ORD-001",
		UserID:      uintPtr(1),
		Status:      status.Pending,
	}
	errCreate := orderRepo.Create(order)
	if errCreate != nil {
		t.Fatalf("Expected no error, got %v", errCreate)
	}

	// Intentar actualizar a estado inválido
	_, err := service.UpdateOrderStatus(order.ID, "INVALID_STATUS")

	if err == nil {
		t.Fatal("Expected error for invalid status")
	}
}

func TestListAllOrders_Success(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear múltiples órdenes
	for i := 1; i <= 5; i++ {
		order := &models.Order{
			OrderNumber: "ORD-" + string(rune(i)),
			UserID:      uintPtr(uint(i % 2)),
			Status:      status.Pending,
			Subtotal:    float64(i * 100),
			Tax:         float64(i * 21),
			Total:       float64(i*100 + i*21),
		}
		errCreate := orderRepo.Create(order)
		if errCreate != nil {
			t.Fatalf("Expected no error, got %v", errCreate)
		}
	}

	// Listar todas las órdenes
	resp, err := service.ListAllOrders(10, 0)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp) != 5 {
		t.Fatalf("Expected 5 orders, got %d", len(resp))
	}
}

func TestGenerateOrderNumber_Unique(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear carrito
	cart := &models.Cart{
		UserID: 1,
		Items: []models.CartItem{
			{
				ProductID: 1,
				Quantity:  1,
				Price:     100.0,
				Product:   &models.Product{ID: 1, Name: "Product 1"},
			},
		},
	}
	cartRepo.carts[1] = cart

	// Crear dos órdenes
	req := &models.CreateOrderRequest{
		ShippingAddress: map[string]interface{}{
			"street": "123 Main St",
		},
	}

	resp1, _ := service.CreateOrder(1, req)

	// Recrear carrito
	cart.Items = []models.CartItem{
		{
			ProductID: 2,
			Quantity:  1,
			Price:     100.0,
			Product:   &models.Product{ID: 2, Name: "Product 2"},
		},
	}

	resp2, _ := service.CreateOrder(1, req)

	// Verificar que los números de orden sean diferentes
	if resp1.OrderNumber == resp2.OrderNumber {
		t.Fatalf("Expected unique order numbers, got %s and %s", resp1.OrderNumber, resp2.OrderNumber)
	}
}

// ============ Benchmark Tests ============

func TestCreateOrder_WithStockReservation(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}
	stockRepo := new(MockProductBranchStockRepository)

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, stockRepo)

	// Setup: Crear carrito con items
	cart := &models.Cart{
		UserID: 1,
		Items: []models.CartItem{
			{
				ProductID: 1,
				Quantity:  2,
				Price:     100.0,
				Product:   &models.Product{ID: 1, Name: "Product 1"},
			},
		},
	}
	cartRepo.carts[1] = cart

	// Crear orden con sucursal origen
	branchID := uint(1)
	req := &models.CreateOrderRequest{
		OriginBranchID: &branchID,
		ShippingAddress: map[string]interface{}{
			"street": "123 Main St",
		},
	}

	// Expectation: stock reservation
	stockRepo.On("IncrementReserved", mock.Anything, uint(1), uint(1), 2).Return(nil)

	_, err := service.CreateOrder(1, req)

	assert.NoError(t, err)
	stockRepo.AssertExpectations(t)
}

func TestCreateGuestOrder_WithStockReservation(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}
	stockRepo := new(MockProductBranchStockRepository)

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, stockRepo)

	// Setup product
	product := &models.Product{
		ID:          1,
		Name:        "Simple Product",
		SKU:         "SIMPLE-123",
		PriceRetail: 100.0,
	}
	variantRepo.On("FindVariantCombinationBySKU", "SIMPLE-123").Return(nil, gorm.ErrRecordNotFound)
	productRepo.On("FindBySKU", "SIMPLE-123").Return(product, nil)

	branchID := uint(1)
	req := &models.CreateGuestOrderRequest{
		GuestEmail:     "guest@example.com",
		OriginBranchID: &branchID,
		Items: []models.GuestCartItem{
			{
				SKU:      "SIMPLE-123",
				Quantity: 2,
				Price:    100.0,
			},
		},
		ShippingAddress: map[string]interface{}{
			"street": "Guest St 123",
		},
	}

	// Expectation: stock reservation
	stockRepo.On("IncrementReserved", mock.Anything, uint(1), uint(1), 2).Return(nil)

	_, err := service.CreateGuestOrder(req)

	assert.NoError(t, err)
	stockRepo.AssertExpectations(t)
}

func BenchmarkCreateOrder(b *testing.B) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	req := &models.CreateOrderRequest{
		ShippingAddress: map[string]interface{}{
			"street": "123 Main St",
			"city":   "Buenos Aires",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := uint(i + 1)

		// Crear carrito con item
		cart := &models.Cart{
			UserID: userID,
			Items: []models.CartItem{
				{
					ProductID: 1,
					Quantity:  2,
					Price:     100.0,
				},
			},
		}
		cartRepo.carts[userID] = cart

		_, err := service.CreateOrder(userID, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetOrderByID(b *testing.B) {
	orderRepo := NewMockOrderRepository()
	cartRepo := NewMockOrderCartRepository()
	userRepo := &MockOrderUserRepository{}
	paymentRepo := &MockPaymentRepository{}
	productRepo := &MockProductRepository{}
	variantRepo := &MockProductVariantRepository{}
	guestRepo := &MockGuestRepository{}

	service := NewOrderService(orderRepo, cartRepo, userRepo, paymentRepo, productRepo, variantRepo, guestRepo, nil)

	// Setup: Crear órdenes
	for i := 1; i <= 100; i++ {
		order := &models.Order{
			OrderNumber: "ORD-" + string(rune(i)),
			UserID:      uintPtr(1),
			Status:      status.Pending,
			Subtotal:    100.0,
			Tax:         21.0,
			Total:       121.0,
		}
		err := orderRepo.Create(order)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderID := uint((i % 100) + 1)
		_, err := service.GetOrderByID(orderID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
