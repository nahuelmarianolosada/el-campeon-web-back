package order

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	orderStatusService "github.com/nahuelmarianolosada/el-campeon-web/internal/services/order/status"
	paymentStatus "github.com/nahuelmarianolosada/el-campeon-web/internal/services/payment/status"
)

type OrderService interface {
	CreateOrder(userID uint, req *models.CreateOrderRequest) (*models.OrderResponse, error)
	GetOrderByID(id uint) (*models.OrderResponse, error)
	GetOrdersByUserID(userID uint, limit, offset int) ([]models.OrderResponse, error)
	UpdateOrderStatus(orderID uint, status string) (*models.OrderResponse, error)
	ListAllOrders(limit, offset int) ([]models.OrderResponse, error)
}

type orderService struct {
	orderRepo   repositories.OrderRepository
	cartRepo    repositories.CartRepository
	userRepo    repositories.UserRepository
	paymentRepo repositories.PaymentRepository
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		userRepo:    userRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *orderService) CreateOrder(userID uint, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// Obtener carrito del usuario
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user's cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Calcular subtotal, tax y total
	subtotal := 0.0
	for _, item := range cart.Items {
		subtotal += float64(item.Quantity) * item.Price
	}

	// Asumir impuesto del 21% (IVA Argentina)
	tax := subtotal * 0.21
	total := subtotal + tax

	// Convertir map a JSON
	shippingData := make(map[string]interface{})
	for k, v := range req.ShippingAddress {
		shippingData[k] = v
	}

	// Crear orden
	order := &models.Order{
		OrderNumber:     s.generateOrderNumber(),
		UserID:          userID,
		Status:          orderStatusService.Pending,
		Subtotal:        subtotal,
		Tax:             tax,
		Total:           total,
		ShippingAddress: shippingData,
		Notes:           req.Notes,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	// Agregar items a la orden
	for _, cartItem := range cart.Items {
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
			Product:   cartItem.Product,
		}
		if err := s.orderRepo.AddItem(order.ID, orderItem); err != nil {
			return nil, fmt.Errorf("error adding item to order: %w", err)
		}
	}

	// Limpiar el carrito
	if err := s.cartRepo.ClearCart(userID); err != nil {
		return nil, fmt.Errorf("error clearing cart: %w", err)
	}

	return s.getOrderResponse(order), nil
}

func (s *orderService) GetOrderByID(id uint) (*models.OrderResponse, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	return s.getOrderResponse(order), nil
}

func (s *orderService) GetOrdersByUserID(userID uint, limit, offset int) ([]models.OrderResponse, error) {
	orders, err := s.orderRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting user's orders: %w", err)
	}

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.getOrderResponse(&order))
	}

	return responses, nil
}

func (s *orderService) UpdateOrderStatus(orderID uint, status string) (*models.OrderResponse, error) {
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	if !orderStatusService.IsValidTransition(order.Status, status) {
		return nil, fmt.Errorf("invalid order status transition: %s -> %s", order.Status, status)
	}

	if err := s.orderRepo.UpdateStatus(orderID, status); err != nil {
		return nil, fmt.Errorf("error updating order status: %w", err)
	}

	paymentFound, err := s.paymentRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("error finding payment for order: %w", err)
	}

	if paymentFound != nil && paymentFound.Status == paymentStatus.Pending {

	}

	return s.GetOrderByID(orderID)
}

func (s *orderService) ListAllOrders(limit, offset int) ([]models.OrderResponse, error) {
	orders, err := s.orderRepo.FindAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing orders: %w", err)
	}

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.getOrderResponse(&order))
	}

	return responses, nil
}

// Helper functions

func (s *orderService) generateOrderNumber() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	prefix := time.Now().Format("20060102")
	randomPart := r.Intn(1000000)
	return fmt.Sprintf("ORD-%s-%06d", prefix, randomPart)
}

func (s *orderService) getOrderResponse(order *models.Order) *models.OrderResponse {
	response := &models.OrderResponse{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		UserID:          order.UserID,
		Status:          order.Status,
		Subtotal:        order.Subtotal,
		Tax:             order.Tax,
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		Notes:           order.Notes,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	// Convertir items
	for _, item := range order.Items {
		response.Items = append(response.Items, models.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Product: models.ProductResponse{
				ID:              item.Product.ID,
				SKU:             item.Product.SKU,
				Name:            item.Product.Name,
				Description:     item.Product.Description,
				Category:        item.Product.Category,
				PriceRetail:     item.Product.PriceRetail,
				PriceWholesale:  item.Product.PriceWholesale,
				Stock:           item.Product.Stock,
				MinBulkQuantity: item.Product.MinBulkQuantity,
				ImageURL:        item.Product.ImageURL,
				IsActive:        item.Product.IsActive,
				CreatedAt:       item.Product.CreatedAt,
			},
			Quantity: item.Quantity,
			Price:    item.Price,
			Subtotal: float64(item.Quantity) * item.Price,
		})
	}

	return response
}
