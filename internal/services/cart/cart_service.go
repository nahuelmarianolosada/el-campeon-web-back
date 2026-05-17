package cart

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type CartService interface {
	AddToCart(userID uint, req *models.AddToCartRequest, isBulkBuyer bool) error
	GetCart(userID uint) (*models.CartResponse, error)
	UpdateCartItem(userID uint, itemID uint, quantity int) error
	RemoveFromCart(userID uint, itemID uint) error
	ClearCart(userID uint) error
	CalculateCartTotal(userID uint) (float64, error)
}

type cartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
	variantRepo repositories.ProductVariantRepository
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository, variantRepo repositories.ProductVariantRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		variantRepo: variantRepo,
	}
}

func (s *cartService) AddToCart(userID uint, req *models.AddToCartRequest, isBulkBuyer bool) error {
	log.Printf("[cartService.AddToCart] INFO: Starting AddToCart for userID=%d, SKU=%s, quantity=%d, isBulkBuyer=%v", userID, req.SKU, req.Quantity, isBulkBuyer)

	// Obtener o crear carrito
	cart, err := s.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		log.Printf("[cartService.AddToCart] ERROR: Failed to get or create cart for userID=%d: %v", userID, err)
		return fmt.Errorf("error getting or creating cart: %w", err)
	}
	log.Printf("[cartService.AddToCart] INFO: Cart retrieved/created successfully - cartID=%d", cart.ID)

	// Intentar obtener como combinación de variante
	variantComb, variantErr := s.variantRepo.FindVariantCombinationBySKU(req.SKU)

	if variantErr == nil && variantComb != nil && variantComb.ID != 0 && variantComb.IsActive {
		log.Printf("[cartService.AddToCart] INFO: Variant combination found - variantID=%d, SKU=%s, stock=%d", variantComb.ID, req.SKU, variantComb.Stock)

		// Es una variante combinación - obtener el producto asociado
		product, err := s.productRepo.FindByID(variantComb.ProductID)
		if err != nil {
			log.Printf("[cartService.AddToCart] ERROR: Failed to find product for variant combination - variantID=%d, productID=%d, error: %v", variantComb.ID, variantComb.ProductID, err)
			return fmt.Errorf("error finding product for variant combination: %w", err)
		}
		log.Printf("[cartService.AddToCart] INFO: Product found for variant - productID=%d, name=%s", product.ID, product.Name)

		// Validar stock de la variante
		if variantComb.Stock < req.Quantity {
			log.Printf("[cartService.AddToCart] WARNING: Insufficient stock for variant combination - variantID=%d, available=%d, requested=%d", variantComb.ID, variantComb.Stock, req.Quantity)
			return fmt.Errorf("insufficient stock for variant combination. available: %d, requested: %d", variantComb.Stock, req.Quantity)
		}

		// Calcular precio: precio base del producto + ajuste de variante
		var price float64
		if isBulkBuyer && req.Quantity >= product.MinBulkQuantity {
			// Para variantes con mayorista, aplicar mayorista al precio base y luego agregar ajuste
			price = product.PriceWholesale + variantComb.PriceAdjustment
			log.Printf("[cartService.AddToCart] INFO: Bulk buyer pricing applied - wholesalePrice=%.2f, priceAdjustment=%.2f, finalPrice=%.2f", product.PriceWholesale, variantComb.PriceAdjustment, price)
		} else {
			price = product.PriceRetail + variantComb.PriceAdjustment
			log.Printf("[cartService.AddToCart] INFO: Retail pricing applied - retailPrice=%.2f, priceAdjustment=%.2f, finalPrice=%.2f", product.PriceRetail, variantComb.PriceAdjustment, price)
		}

		// Crear item del carrito con referencia a la variante combinación
		item := &models.CartItem{
			CartID:                      cart.ID,
			ProductID:                   product.ID,
			ProductVariantCombinationID: &variantComb.ID,
			Quantity:                    req.Quantity,
			Price:                       price,
		}

		if err := s.cartRepo.AddItem(userID, item); err != nil {
			log.Printf("[cartService.AddToCart] ERROR: Failed to add variant combination to cart - userID=%d, variantID=%d, error: %v", userID, variantComb.ID, err)
			return fmt.Errorf("error adding variant combination to cart: %w", err)
		}
		log.Printf("[cartService.AddToCart] INFO: Variant combination successfully added to cart - userID=%d, variantID=%d, quantity=%d, price=%.2f", userID, variantComb.ID, req.Quantity, price)

		return nil
	}

	if variantErr != nil {
		log.Printf("[cartService.AddToCart] INFO: Variant combination not found (searching for standard product) - SKU=%s, error: %v", req.SKU, variantErr)
	}

	// Si no es variante, buscar producto normal
	product, err := s.productRepo.FindBySKU(req.SKU)
	if err != nil && err.Error() != "record not found" {
		log.Printf("[cartService.AddToCart] ERROR: Failed to find product by SKU - SKU=%s, error: %v", req.SKU, err)
		return fmt.Errorf("error finding product: %w", err)
	}

	if product == nil || product.ID == 0 {
		log.Printf("[cartService.AddToCart] ERROR: Product not found - SKU=%s, no variant combination found either", req.SKU)
		return fmt.Errorf("product or variant combination not found with SKU: %s", req.SKU)
	}
	log.Printf("[cartService.AddToCart] INFO: Standard product found - productID=%d, name=%s, stock=%d", product.ID, product.Name, product.Stock)

	if product.Stock < req.Quantity {
		log.Printf("[cartService.AddToCart] WARNING: Insufficient stock for product - productID=%d, available=%d, requested=%d", product.ID, product.Stock, req.Quantity)
		return fmt.Errorf("insufficient stock. available: %d, requested: %d", product.Stock, req.Quantity)
	}

	// Determinar precio basado en si es comprador mayorista
	var price float64
	if isBulkBuyer && req.Quantity >= product.MinBulkQuantity {
		price = product.PriceWholesale
		log.Printf("[cartService.AddToCart] INFO: Bulk buyer pricing applied - productID=%d, wholesalePrice=%.2f, quantity=%d", product.ID, price, req.Quantity)
	} else {
		price = product.PriceRetail
		log.Printf("[cartService.AddToCart] INFO: Retail pricing applied - productID=%d, retailPrice=%.2f", product.ID, price)
	}

	// Crear item del carrito sin referencia a variante
	item := &models.CartItem{
		CartID:    cart.ID,
		ProductID: product.ID,
		Quantity:  req.Quantity,
		Price:     price,
	}

	if err := s.cartRepo.AddItem(userID, item); err != nil {
		log.Printf("[cartService.AddToCart] ERROR: Failed to add item to cart - userID=%d, productID=%d, error: %v", userID, product.ID, err)
		return fmt.Errorf("error adding item to cart: %w", err)
	}
	log.Printf("[cartService.AddToCart] INFO: Product successfully added to cart - userID=%d, productID=%d, quantity=%d, price=%.2f", userID, product.ID, req.Quantity, price)

	return nil
}

func (s *cartService) GetCart(userID uint) (*models.CartResponse, error) {
	log.Printf("[cartService.GetCart] INFO: Retrieving cart for userID=%d", userID)

	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		log.Printf("[cartService.GetCart] ERROR: Failed to retrieve cart for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error getting cart: %w", err)
	}
	log.Printf("[cartService.GetCart] INFO: Cart retrieved successfully - cartID=%d, itemCount=%d", cart.ID, len(cart.Items))

	// Convertir a response
	total := s.calculateTotal(cart.Items)
	log.Printf("[cartService.GetCart] INFO: Cart total calculated - total=%.2f", total)

	return &models.CartResponse{
		ID:     cart.ID,
		UserID: cart.UserID,
		Items:  s.toCartItemResponses(cart.Items),
		Total:  total,
	}, nil
}

func (s *cartService) UpdateCartItem(userID uint, itemID uint, quantity int) error {
	log.Printf("[cartService.UpdateCartItem] INFO: Starting update of cart item - userID=%d, itemID=%d, newQuantity=%d", userID, itemID, quantity)

	// Obtener el item actual
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		log.Printf("[cartService.UpdateCartItem] ERROR: Failed to retrieve cart - userID=%d, error: %v", userID, err)
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
		log.Printf("[cartService.UpdateCartItem] ERROR: Cart item not found - userID=%d, itemID=%d", userID, itemID)
		return fmt.Errorf("cart item not found")
	}
	log.Printf("[cartService.UpdateCartItem] INFO: Cart item found - itemID=%d, productID=%d, currentQuantity=%d", item.ID, item.ProductID, item.Quantity)

	// Validar stock disponible
	product, err := s.productRepo.FindByID(item.ProductID)
	if err != nil {
		log.Printf("[cartService.UpdateCartItem] ERROR: Failed to find product - productID=%d, error: %v", item.ProductID, err)
		return fmt.Errorf("error finding product: %w", err)
	}

	if product.Stock < quantity {
		log.Printf("[cartService.UpdateCartItem] WARNING: Insufficient stock for update - productID=%d, available=%d, requested=%d", product.ID, product.Stock, quantity)
		return fmt.Errorf("insufficient stock. available: %d, requested: %d", product.Stock, quantity)
	}

	// Actualizar precio si cambió el estatus de mayorista
	var price float64
	if quantity >= product.MinBulkQuantity {
		price = product.PriceWholesale
		log.Printf("[cartService.UpdateCartItem] INFO: Bulk pricing threshold reached - quantity=%d, minBulkQuantity=%d, newPrice=%.2f", quantity, product.MinBulkQuantity, price)
	} else {
		price = product.PriceRetail
		log.Printf("[cartService.UpdateCartItem] INFO: Retail pricing applied - newPrice=%.2f", price)
	}

	item.Price = price
	item.Quantity = quantity
	item.UpdatedAt = time.Now()

	if err := s.cartRepo.UpdateCompleteItemInTheCart(item); err != nil {
		log.Printf("[cartService.UpdateCartItem] ERROR: Failed to update cart item in repository - itemID=%d, error: %v", itemID, err)
		return fmt.Errorf("error updating cart item: %w", err)
	}
	log.Printf("[cartService.UpdateCartItem] INFO: Cart item successfully updated - itemID=%d, newQuantity=%d, newPrice=%.2f", itemID, quantity, price)

	return nil
}

func (s *cartService) RemoveFromCart(userID uint, itemID uint) error {
	log.Printf("[cartService.RemoveFromCart] INFO: Starting removal of cart item - userID=%d, itemID=%d", userID, itemID)

	// Validar que el item pertenece al carrito del usuario
	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		log.Printf("[cartService.RemoveFromCart] ERROR: Failed to retrieve cart - userID=%d, error: %v", userID, err)
		return fmt.Errorf("error getting cart: %w", err)
	}

	found := false
	for _, item := range cart.Items {
		if item.ID == itemID {
			found = true
			log.Printf("[cartService.RemoveFromCart] INFO: Item found in cart - itemID=%d, productID=%d", itemID, item.ProductID)
			break
		}
	}

	if !found {
		log.Printf("[cartService.RemoveFromCart] ERROR: Item not found in user's cart - userID=%d, itemID=%d", userID, itemID)
		return fmt.Errorf("item not found in user's cart")
	}

	if err := s.cartRepo.RemoveItem(itemID); err != nil {
		log.Printf("[cartService.RemoveFromCart] ERROR: Failed to remove item from repository - itemID=%d, error: %v", itemID, err)
		return fmt.Errorf("error removing item: %w", err)
	}
	log.Printf("[cartService.RemoveFromCart] INFO: Item successfully removed from cart - userID=%d, itemID=%d", userID, itemID)

	return nil
}

func (s *cartService) ClearCart(userID uint) error {
	log.Printf("[cartService.ClearCart] INFO: Starting clear cart operation - userID=%d", userID)

	if err := s.cartRepo.ClearCart(userID); err != nil {
		log.Printf("[cartService.ClearCart] ERROR: Failed to clear cart - userID=%d, error: %v", userID, err)
		return fmt.Errorf("error clearing cart: %w", err)
	}
	log.Printf("[cartService.ClearCart] INFO: Cart successfully cleared - userID=%d", userID)

	return nil
}

func (s *cartService) CalculateCartTotal(userID uint) (float64, error) {
	log.Printf("[cartService.CalculateCartTotal] INFO: Starting cart total calculation - userID=%d", userID)

	cart, err := s.cartRepo.GetCart(userID)
	if err != nil {
		log.Printf("[cartService.CalculateCartTotal] ERROR: Failed to retrieve cart - userID=%d, error: %v", userID, err)
		return 0, fmt.Errorf("error getting cart: %w", err)
	}

	total := s.calculateTotal(cart.Items)
	log.Printf("[cartService.CalculateCartTotal] INFO: Cart total calculated successfully - userID=%d, itemCount=%d, total=%.2f", userID, len(cart.Items), total)

	return total, nil
}

// Helper functions

func (s *cartService) calculateTotal(items []models.CartItem) float64 {
	total := 0.0
	for _, item := range items {
		itemSubtotal := float64(item.Quantity) * item.Price
		total += itemSubtotal
		log.Printf("[cartService.calculateTotal] DEBUG: Item calculation - itemID=%d, quantity=%d, price=%.2f, subtotal=%.2f", item.ID, item.Quantity, item.Price, itemSubtotal)
	}
	log.Printf("[cartService.calculateTotal] INFO: Total calculated - itemCount=%d, total=%.2f", len(items), total)
	return total
}

func (s *cartService) toCartItemResponses(items []models.CartItem) []models.CartItemResponse {
	log.Printf("[cartService.toCartItemResponses] INFO: Converting cart items to responses - itemCount=%d", len(items))

	var responses []models.CartItemResponse
	for _, item := range items {
		itemResponse := models.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Product:   *s.toProductResponse(item.Product),
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  float64(item.Quantity) * item.Price,
		}

		// Agregar información de variante si existe
		if item.ProductVariantCombination != nil {
			itemResponse.VariantCombination = &models.ProductVariantCombinationResponse{
				ID:                 item.ProductVariantCombination.ID,
				SKU:                item.ProductVariantCombination.SKU,
				VariantCombination: s.parseVariantCombination(item.ProductVariantCombination),
				Stock:              item.ProductVariantCombination.Stock,
				PriceAdjustment:    item.ProductVariantCombination.PriceAdjustment,
				ImageURL:           item.ProductVariantCombination.ImageURL,
				IsActive:           item.ProductVariantCombination.IsActive,
				CreatedAt:          item.ProductVariantCombination.CreatedAt,
			}
			log.Printf("[cartService.toCartItemResponses] INFO: Variant combination added to response - itemID=%d, variantID=%d", item.ID, item.ProductVariantCombination.ID)
		}

		responses = append(responses, itemResponse)
	}
	log.Printf("[cartService.toCartItemResponses] INFO: Successfully converted %d items to responses", len(responses))
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
		ImageURL:        models.PrimaryImageURL(product.Images),
		Images:          models.ToProductImageResponses(product.Images),
		IsActive:        product.IsActive,
		CreatedAt:       product.CreatedAt,
	}
}

// parseVariantCombination convierte el JSON serializado de variantes a un mapa
func (s *cartService) parseVariantCombination(variant *models.ProductVariantCombination) map[string]string {
	if variant == nil || variant.VariantCombination == "" {
		log.Printf("[cartService.parseVariantCombination] WARNING: Variant combination is empty or nil")
		return nil
	}

	var combination map[string]string
	if err := json.Unmarshal([]byte(variant.VariantCombination), &combination); err != nil {
		log.Printf("[cartService.parseVariantCombination] ERROR: Failed to unmarshal variant combination JSON - variantID=%d, error: %v", variant.ID, err)
		// Si no se puede parsear, retornar nil
		return nil
	}
	log.Printf("[cartService.parseVariantCombination] INFO: Successfully parsed variant combination - variantID=%d, attributeCount=%d", variant.ID, len(combination))
	return combination
}

