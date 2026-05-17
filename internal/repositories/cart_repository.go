package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetOrCreateCart(userID uint) (*models.Cart, error)
	AddItem(userID uint, item *models.CartItem) error
	UpdateItem(itemID uint, quantity int) error
	UpdateCompleteItemInTheCart(cartItem *models.CartItem) error
	RemoveItem(itemID uint) error
	GetCart(userID uint) (*models.Cart, error)
	ClearCart(userID uint) error
	GetCartItems(userID uint) ([]models.CartItem, error)
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreateCart(userID uint) (*models.Cart, error) {
	log.Printf("[cartRepository.GetOrCreateCart] INFO: Getting or creating cart - userID=%d", userID)

	var cart models.Cart
	result := r.db.First(&cart, "user_id = ?", userID)
	if result.Error == gorm.ErrRecordNotFound {
		// Crear nuevo carrito
		log.Printf("[cartRepository.GetOrCreateCart] INFO: Cart not found, creating new cart - userID=%d", userID)
		cart = models.Cart{UserID: userID}
		if err := r.db.Create(&cart).Error; err != nil {
			log.Printf("[cartRepository.GetOrCreateCart] ERROR: Failed to create cart - userID=%d: %v", userID, err)
			return nil, err
		}
		log.Printf("[cartRepository.GetOrCreateCart] INFO: New cart created successfully - cartID=%d, userID=%d", cart.ID, userID)
	} else if result.Error != nil {
		log.Printf("[cartRepository.GetOrCreateCart] ERROR: Failed to retrieve cart - userID=%d: %v", userID, result.Error)
		return nil, result.Error
	} else {
		log.Printf("[cartRepository.GetOrCreateCart] INFO: Existing cart found - cartID=%d, userID=%d", cart.ID, userID)
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(userID uint, item *models.CartItem) error {
	log.Printf("[cartRepository.AddItem] INFO: Adding item to cart - userID=%d, productID=%d, quantity=%d", userID, item.ProductID, item.Quantity)

	cart, err := r.GetOrCreateCart(userID)
	if err != nil {
		log.Printf("[cartRepository.AddItem] ERROR: Failed to get or create cart - userID=%d: %v", userID, err)
		return err
	}

	item.CartID = cart.ID
	if err := r.db.Create(item).Error; err != nil {
		log.Printf("[cartRepository.AddItem] ERROR: Failed to add item to cart - cartID=%d, productID=%d: %v", cart.ID, item.ProductID, err)
		return err
	}
	log.Printf("[cartRepository.AddItem] INFO: Item added to cart successfully - cartID=%d, itemID=%d, productID=%d", cart.ID, item.ID, item.ProductID)
	return nil
}

func (r *cartRepository) UpdateItem(itemID uint, quantity int) error {
	log.Printf("[cartRepository.UpdateItem] INFO: Updating cart item quantity - itemID=%d, newQuantity=%d", itemID, quantity)

	if err := r.db.Model(&models.CartItem{}).Where("id = ?", itemID).Update("quantity", quantity).Error; err != nil {
		log.Printf("[cartRepository.UpdateItem] ERROR: Failed to update item - itemID=%d: %v", itemID, err)
		return err
	}
	log.Printf("[cartRepository.UpdateItem] INFO: Item updated successfully - itemID=%d, quantity=%d", itemID, quantity)
	return nil
}

func (r *cartRepository) UpdateCompleteItemInTheCart(cartItem *models.CartItem) error {
	log.Printf("[cartRepository.UpdateCompleteItemInTheCart] INFO: Updating complete cart item - itemID=%d, quantity=%d, price=%.2f", cartItem.ID, cartItem.Quantity, cartItem.Price)

	if err := r.db.Save(cartItem).Error; err != nil {
		log.Printf("[cartRepository.UpdateCompleteItemInTheCart] ERROR: Failed to update item - itemID=%d: %v", cartItem.ID, err)
		return err
	}
	log.Printf("[cartRepository.UpdateCompleteItemInTheCart] INFO: Cart item updated successfully - itemID=%d", cartItem.ID)
	return nil
}

func (r *cartRepository) RemoveItem(itemID uint) error {
	log.Printf("[cartRepository.RemoveItem] INFO: Removing item from cart - itemID=%d", itemID)

	if err := r.db.Delete(&models.CartItem{}, itemID).Error; err != nil {
		log.Printf("[cartRepository.RemoveItem] ERROR: Failed to remove item - itemID=%d: %v", itemID, err)
		return err
	}
	log.Printf("[cartRepository.RemoveItem] INFO: Item removed successfully - itemID=%d", itemID)
	return nil
}

func (r *cartRepository) GetCart(userID uint) (*models.Cart, error) {
	log.Printf("[cartRepository.GetCart] INFO: Retrieving cart with items - userID=%d", userID)

	var cart models.Cart
	if err := r.db.Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images", orderImages).
		Preload("Items.ProductVariantCombination").
		Preload("User").
		First(&cart, "user_id = ?", userID).
		Error; err != nil {
		log.Printf("[cartRepository.GetCart] ERROR: Failed to retrieve cart - userID=%d: %v", userID, err)
		return nil, err
	}
	log.Printf("[cartRepository.GetCart] INFO: Cart retrieved with items - cartID=%d, userID=%d, itemCount=%d", cart.ID, userID, len(cart.Items))
	return &cart, nil
}

func (r *cartRepository) ClearCart(userID uint) error {
	log.Printf("[cartRepository.ClearCart] INFO: Clearing cart - userID=%d", userID)

	var cart models.Cart
	if err := r.db.First(&cart, "user_id = ?", userID).Error; err != nil {
		log.Printf("[cartRepository.ClearCart] ERROR: Failed to find cart - userID=%d: %v", userID, err)
		return err
	}

	if err := r.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		log.Printf("[cartRepository.ClearCart] ERROR: Failed to delete cart items - cartID=%d: %v", cart.ID, err)
		return err
	}
	log.Printf("[cartRepository.ClearCart] INFO: Cart cleared successfully - cartID=%d, userID=%d", cart.ID, userID)
	return nil
}

func (r *cartRepository) GetCartItems(userID uint) ([]models.CartItem, error) {
	log.Printf("[cartRepository.GetCartItems] INFO: Retrieving cart items - userID=%d", userID)

	var cart models.Cart
	if err := r.db.First(&cart, "user_id = ?", userID).Error; err != nil {
		log.Printf("[cartRepository.GetCartItems] ERROR: Failed to find cart - userID=%d: %v", userID, err)
		return nil, err
	}

	var items []models.CartItem
	if err := r.db.Preload("Product").Preload("Product.Images", orderImages).Where("cart_id = ?", cart.ID).Find(&items).Error; err != nil {
		log.Printf("[cartRepository.GetCartItems] ERROR: Failed to retrieve items - cartID=%d: %v", cart.ID, err)
		return nil, err
	}
	log.Printf("[cartRepository.GetCartItems] INFO: Cart items retrieved - cartID=%d, itemCount=%d", cart.ID, len(items))
	return items, nil
}
