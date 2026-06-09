package order

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	orderStatusService "github.com/nahuelmarianolosada/el-campeon-web/internal/services/order/status"
)

type OrderService interface {
	CreateOrder(userID uint, req *models.CreateOrderRequest) (*models.OrderResponse, error)
	CreateGuestOrder(req *models.CreateGuestOrderRequest) (*models.OrderResponse, error)
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
	productRepo repositories.ProductRepository
	variantRepo repositories.ProductVariantRepository
	guestRepo   repositories.GuestRepository
	stockRepo   repositories.ProductBranchStockRepository
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
	productRepo repositories.ProductRepository,
	variantRepo repositories.ProductVariantRepository,
	guestRepo repositories.GuestRepository,
	stockRepo repositories.ProductBranchStockRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		userRepo:    userRepo,
		paymentRepo: paymentRepo,
		productRepo: productRepo,
		variantRepo: variantRepo,
		guestRepo:   guestRepo,
		stockRepo:   stockRepo,
	}
}

func (s *orderService) CreateOrder(userID uint, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	log.Printf("[orderService.CreateOrder] INFO: Starting order creation for userID=%d, deliveryMethod=%s", userID, req.DeliveryMethod)

	// Obtener carrito del usuario
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		log.Printf("[orderService.CreateOrder] ERROR: Failed to get cart for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error getting user's cart: %w", err)
	}

	if len(cart.Items) == 0 {
		log.Printf("[orderService.CreateOrder] WARNING: Cart is empty for userID=%d", userID)
		return nil, fmt.Errorf("cart is empty")
	}
	log.Printf("[orderService.CreateOrder] INFO: Cart retrieved successfully - cartID=%d, itemCount=%d", cart.ID, len(cart.Items))

	// Calcular subtotal, tax y total
	subtotal := 0.0
	for _, item := range cart.Items {
		subtotal += float64(item.Quantity) * item.Price
	}

	// Asumir impuesto del 21% (IVA Argentina)
	tax := subtotal * 0.21
	total := subtotal + tax
	log.Printf("[orderService.CreateOrder] INFO: Calculations completed - subtotal=%.2f, tax=%.2f, total=%.2f", subtotal, tax, total)

	// Convertir map a JSON
	shippingData := make(map[string]interface{})
	for k, v := range req.ShippingAddress {
		shippingData[k] = v
	}

	// Total incluye el costo de envío que el front cotizó previamente (snapshot).
	total += req.ShippingCost

	// Crear orden
	orderNumber := s.generateOrderNumber()
	order := &models.Order{
		OrderNumber:     orderNumber,
		UserID:          &userID,
		Status:          orderStatusService.Pending,
		Subtotal:        subtotal,
		Tax:             tax,
		Total:           total,
		ShippingCost:    req.ShippingCost,
		ShippingAddress: shippingData,
		DeliveryMethod:  req.DeliveryMethod,
		OriginBranchID:  req.OriginBranchID,
		DeliveryZoneID:  req.DeliveryZoneID,
		Notes:           req.Notes,
	}

	if err := s.orderRepo.Create(order); err != nil {
		log.Printf("[orderService.CreateOrder] ERROR: Failed to create order for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error creating order: %w", err)
	}
	log.Printf("[orderService.CreateOrder] INFO: Order created successfully - orderID=%d, orderNumber=%s", order.ID, orderNumber)

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
			log.Printf("[orderService.CreateOrder] ERROR: Failed to add item to order - orderID=%d, productID=%d: %v", order.ID, cartItem.ProductID, err)
			return nil, fmt.Errorf("error adding item to order: %w", err)
		}
	}
	log.Printf("[orderService.CreateOrder] INFO: All items added to order - orderID=%d, itemCount=%d", order.ID, len(cart.Items))

	// Reservar stock en la sucursal origen si se especificó.
	if order.OriginBranchID != nil && s.stockRepo != nil {
		for _, item := range cart.Items {
			if err := s.stockRepo.IncrementReserved(nil, item.ProductID, *order.OriginBranchID, item.Quantity); err != nil {
				log.Printf("[orderService.CreateOrder] WARN: Failed to reserve stock - productID=%d, branchID=%d, qty=%d: %v", item.ProductID, *order.OriginBranchID, item.Quantity, err)
			}
		}
	}

	// Limpiar el carrito
	if err := s.cartRepo.ClearCart(userID); err != nil {
		log.Printf("[orderService.CreateOrder] ERROR: Failed to clear cart for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error clearing cart: %w", err)
	}
	log.Printf("[orderService.CreateOrder] INFO: Cart cleared successfully for userID=%d", userID)

	return s.getOrderResponse(order), nil
}

func (s *orderService) GetOrderByID(id uint) (*models.OrderResponse, error) {
	log.Printf("[orderService.GetOrderByID] INFO: Retrieving order - orderID=%d", id)

	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		log.Printf("[orderService.GetOrderByID] ERROR: Failed to find order - orderID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding order: %w", err)
	}
	log.Printf("[orderService.GetOrderByID] INFO: Order found successfully - orderID=%d, orderNumber=%s, status=%s", order.ID, order.OrderNumber, order.Status)

	return s.getOrderResponse(order), nil
}

func (s *orderService) GetOrdersByUserID(userID uint, limit, offset int) ([]models.OrderResponse, error) {
	log.Printf("[orderService.GetOrdersByUserID] INFO: Retrieving orders for user - userID=%d, limit=%d, offset=%d", userID, limit, offset)

	orders, err := s.orderRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		log.Printf("[orderService.GetOrdersByUserID] ERROR: Failed to get orders for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error getting user's orders: %w", err)
	}
	log.Printf("[orderService.GetOrdersByUserID] INFO: Orders retrieved successfully - userID=%d, orderCount=%d", userID, len(orders))

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.getOrderResponse(&order))
	}

	return responses, nil
}

func (s *orderService) UpdateOrderStatus(orderID uint, status string) (*models.OrderResponse, error) {
	log.Printf("[orderService.UpdateOrderStatus] INFO: Starting order status update - orderID=%d, newStatus=%s", orderID, status)

	order, err := s.GetOrderByID(orderID)
	if err != nil {
		log.Printf("[orderService.UpdateOrderStatus] ERROR: Failed to find order - orderID=%d: %v", orderID, err)
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	if !orderStatusService.IsValidTransition(order.Status, status) {
		log.Printf("[orderService.UpdateOrderStatus] WARNING: Invalid status transition - orderID=%d, currentStatus=%s, requestedStatus=%s", orderID, order.Status, status)
		return nil, fmt.Errorf("invalid order status transition: %s -> %s", order.Status, status)
	}

	if err := s.orderRepo.UpdateStatus(orderID, status); err != nil {
		log.Printf("[orderService.UpdateOrderStatus] ERROR: Failed to update status - orderID=%d, newStatus=%s: %v", orderID, status, err)
		return nil, fmt.Errorf("error updating order status: %w", err)
	}
	log.Printf("[orderService.UpdateOrderStatus] INFO: Order status updated successfully - orderID=%d, oldStatus=%s, newStatus=%s", orderID, order.Status, status)

	return s.GetOrderByID(orderID)
}

func (s *orderService) ListAllOrders(limit, offset int) ([]models.OrderResponse, error) {
	log.Printf("[orderService.ListAllOrders] INFO: Listing all orders - limit=%d, offset=%d", limit, offset)

	orders, err := s.orderRepo.FindAll(limit, offset)
	if err != nil {
		log.Printf("[orderService.ListAllOrders] ERROR: Failed to list orders: %v", err)
		return nil, fmt.Errorf("error listing orders: %w", err)
	}
	log.Printf("[orderService.ListAllOrders] INFO: Orders listed successfully - orderCount=%d", len(orders))

	var responses []models.OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.getOrderResponse(&order))
	}

	return responses, nil
}

// Helper functions

func (s *orderService) generateOrderNumber() string {
	log.Printf("[orderService.generateOrderNumber] INFO: Generating new order number")
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
		GuestEmail:      order.GuestEmail,
		Status:          order.Status,
		Subtotal:        order.Subtotal,
		Tax:             order.Tax,
		Total:           order.Total,
		ShippingCost:    order.ShippingCost,
		ShippingAddress: order.ShippingAddress,
		DeliveryMethod:  order.DeliveryMethod,
		OriginBranchID:  order.OriginBranchID,
		DeliveryZoneID:  order.DeliveryZoneID,
		Notes:           order.Notes,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	// Establecer UserID si no es nulo
	if order.UserID != nil {
		response.UserID = *order.UserID
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
				ImageURL:        models.PrimaryImageURL(item.Product.Images),
				Images:          models.ToProductImageResponses(item.Product.Images),
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

// CreateGuestOrder crea una orden desde un carrito guest
func (s *orderService) CreateGuestOrder(req *models.CreateGuestOrderRequest) (*models.OrderResponse, error) {
	log.Printf("[orderService.CreateGuestOrder] INFO: Starting guest order creation - email=%s, itemCount=%d", req.GuestEmail, len(req.Items))

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Validar SKUs y precios, acumular subtotal
	subtotal := 0.0
	var resolvedItems []models.OrderItem

	for _, cartItem := range req.Items {
		// Intentar variante primero
		variant, err := s.variantRepo.FindVariantCombinationBySKU(cartItem.SKU)
		if err == nil && variant != nil {
			product, err := s.productRepo.FindByID(variant.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product not found for item %s", cartItem.SKU)
			}
			// Calcular precio esperado: precio base + ajuste de variante
			expectedPrice := product.PriceRetail + variant.PriceAdjustment
			if cartItem.Price != expectedPrice {
				return nil, fmt.Errorf("price mismatch for item %s: expected %.2f, got %.2f", cartItem.SKU, expectedPrice, cartItem.Price)
			}
			subtotal += float64(cartItem.Quantity) * cartItem.Price
			resolvedItems = append(resolvedItems, models.OrderItem{
				ProductID: product.ID,
				Quantity:  cartItem.Quantity,
				Price:     cartItem.Price,
				Product:   product,
			})
		} else {
			// Intentar producto simple
			product, err := s.productRepo.FindBySKU(cartItem.SKU)
			if err != nil {
				return nil, fmt.Errorf("product not found: %s", cartItem.SKU)
			}
			if cartItem.Price != product.PriceRetail {
				return nil, fmt.Errorf("price mismatch for item %s: expected %.2f, got %.2f", cartItem.SKU, product.PriceRetail, cartItem.Price)
			}
			subtotal += float64(cartItem.Quantity) * cartItem.Price
			resolvedItems = append(resolvedItems, models.OrderItem{
				ProductID: product.ID,
				Quantity:  cartItem.Quantity,
				Price:     cartItem.Price,
				Product:   product,
			})
		}
	}

	// Calcular tax y total (incluyendo costo de envío cotizado)
	tax := subtotal * 0.21
	total := subtotal + tax + req.ShippingCost

	// Convertir map a JSON
	shippingData := make(map[string]interface{})
	for k, v := range req.ShippingAddress {
		shippingData[k] = v
	}

	// Crear orden
	orderNumber := s.generateOrderNumber()
	order := &models.Order{
		OrderNumber:     orderNumber,
		GuestEmail:      req.GuestEmail,
		Status:          orderStatusService.Pending,
		Subtotal:        subtotal,
		Tax:             tax,
		Total:           total,
		ShippingCost:    req.ShippingCost,
		ShippingAddress: shippingData,
		DeliveryMethod:  req.DeliveryMethod,
		OriginBranchID:  req.OriginBranchID,
		DeliveryZoneID:  req.DeliveryZoneID,
		Notes:           req.Notes,
	}

	if req.UserID != 0 {
		order.UserID = &req.UserID
	}

	if err := s.orderRepo.Create(order); err != nil {
		log.Printf("[orderService.CreateGuestOrder] ERROR: Failed to create order: %v", err)
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	// Agregar items
	for _, item := range resolvedItems {
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Product:   item.Product,
		}
		if err := s.orderRepo.AddItem(order.ID, orderItem); err != nil {
			return nil, fmt.Errorf("error adding item to order: %w", err)
		}
	}

	order.Items = resolvedItems

	// Reservar stock en la sucursal origen (si aplica)
	if order.OriginBranchID != nil && s.stockRepo != nil {
		for _, item := range resolvedItems {
			if err := s.stockRepo.IncrementReserved(nil, item.ProductID, *order.OriginBranchID, item.Quantity); err != nil {
				log.Printf("[orderService.CreateGuestOrder] WARN: Failed to reserve stock - productID=%d, branchID=%d, qty=%d: %v", item.ProductID, *order.OriginBranchID, item.Quantity, err)
			}
		}
	}

	log.Printf("[orderService.CreateGuestOrder] INFO: Guest order created - orderID=%d", order.ID)
	return s.getOrderResponse(order), nil
}
