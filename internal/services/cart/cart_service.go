package cart

import (
	"fmt"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type CartService interface {
	AddToCart(userID uint, req *models.AddToCartRequest, isBulkBuyer bool) error
	GetCart(userID uint) (*models.CartResponse, error)
	UpdateCartItem(userID uint, itemID uint, quantity int, isBulkBuyer bool) error
	RemoveFromCart(userID uint, itemID uint) error
	ClearCart(userID uint) error
	CalculateCartTotal(userID uint) (float64, error)
}

type cartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) AddToCart(userID uint, req *models.AddToCartRequest, isBulkBuyer bool) error {
	// Obtener o crear carrito
	cart, err := s.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return fmt.Errorf("error getting or creating cart: %w", err)
	}

	// Validar que el producto existe y tiene stock
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return fmt.Errorf("error finding product: %w", err)
	}

	if product.Stock < req.Quantity {
		return fmt.Errorf("insufficient stock. available: %d, requested: %d", product.Stock, req.Quantity)
	}

	// Determinar precio basado en si es comprador mayorista
	var price float64
	if isBulkBuyer && req.Quantity >= product.MinBulkQuantity {
		price = product.PriceWholesale
	} else {
		price = product.PriceRetail
	}

	// Crear item del carrito
	item := &models.CartItem{
		CartID:    cart.ID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     price,
	}

	if err := s.cartRepo.AddItem(userID, item); err != nil {
		return fmt.Errorf("error adding item to cart: %w", err)
	}

	return nil
}

func (s *cartService) GetCart(userID uint) (*models.CartResponse, error) {
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// Convertir a response
	total := s.calculateTotal(cart.Items)
	return &models.CartResponse{
		ID:     cart.ID,
		UserID: cart.UserID,
		Items:  s.toCartItemResponses(cart.Items),
		Total:  total,
	}, nil
}

func (s *cartService) UpdateCartItem(userID uint, itemID uint, quantity int, isBulkBuyer bool) error {
	// Obtener el item actual
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		return fmt.Errorf("error getting cart: %w", err)
	}

	var item *models.CartItem
	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			item = &cart.Items[i]
			break
		}
	}

	if item == nil {
		return fmt.Errorf("cart item not found")
	}

	// Validar stock disponible
	product, err := s.productRepo.FindByID(item.ProductID)
	if err != nil {
		return fmt.Errorf("error finding product: %w", err)
	}

	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock. available: %d, requested: %d", product.Stock, quantity)
	}

	// Actualizar precio si cambió el estatus de mayorista
	/*var price float64
	if isBulkBuyer && quantity >= product.MinBulkQuantity {
		price = product.PriceWholesale
	} else {
		price = product.PriceRetail
	}*/

	// Actualizar la cantidad en la BD
	if err := s.cartRepo.UpdateItem(itemID, quantity); err != nil {
		return fmt.Errorf("error updating cart item: %w", err)
	}

	// Si el precio cambió, necesitamos actualizar también
	// Por ahora solo actualizamos la cantidad
	return nil
}

func (s *cartService) RemoveFromCart(userID uint, itemID uint) error {
	// Validar que el item pertenece al carrito del usuario
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		return fmt.Errorf("error getting cart: %w", err)
	}

	found := false
	for _, item := range cart.Items {
		if item.ID == itemID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item not found in user's cart")
	}

	if err := s.cartRepo.RemoveItem(itemID); err != nil {
		return fmt.Errorf("error removing item: %w", err)
	}

	return nil
}

func (s *cartService) ClearCart(userID uint) error {
	if err := s.cartRepo.ClearCart(userID); err != nil {
		return fmt.Errorf("error clearing cart: %w", err)
	}
	return nil
}

func (s *cartService) CalculateCartTotal(userID uint) (float64, error) {
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		return 0, fmt.Errorf("error getting cart: %w", err)
	}

	return s.calculateTotal(cart.Items), nil
}

// Helper functions

func (s *cartService) calculateTotal(items []models.CartItem) float64 {
	total := 0.0
	for _, item := range items {
		total += float64(item.Quantity) * item.Price
	}
	return total
}

func (s *cartService) toCartItemResponses(items []models.CartItem) []models.CartItemResponse {
	var responses []models.CartItemResponse
	for _, item := range items {
		responses = append(responses, models.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Product:   *s.toProductResponse(item.Product),
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  float64(item.Quantity) * item.Price,
		})
	}
	return responses
}

func (s *cartService) toProductResponse(product *models.Product) *models.ProductResponse {
	return &models.ProductResponse{
		ID:              product.ID,
		SKU:             product.SKU,
		Name:            product.Name,
		Description:     product.Description,
		Category:        product.Category,
		PriceRetail:     product.PriceRetail,
		PriceWholesale:  product.PriceWholesale,
		Stock:           product.Stock,
		MinBulkQuantity: product.MinBulkQuantity,
		ImageURL:        product.ImageURL,
		IsActive:        product.IsActive,
		CreatedAt:       product.CreatedAt,
	}
}
