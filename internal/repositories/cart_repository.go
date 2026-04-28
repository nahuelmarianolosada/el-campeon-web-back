package repositories

import (
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetOrCreateCart(userID uint) (*models.Cart, error)
	AddItem(userID uint, item *models.CartItem) error
	UpdateItem(itemID uint, quantity int) error
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
	var cart models.Cart
	result := r.db.First(&cart, "user_id = ?", userID)
	if result.Error == gorm.ErrRecordNotFound {
		// Crear nuevo carrito
		cart = models.Cart{UserID: userID}
		if err := r.db.Create(&cart).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(userID uint, item *models.CartItem) error {
	cart, err := r.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	item.CartID = cart.ID
	return r.db.Create(item).Error
}

func (r *cartRepository) UpdateItem(itemID uint, quantity int) error {
	return r.db.Model(&models.CartItem{}).Where("id = ?", itemID).Update("quantity", quantity).Error
}

func (r *cartRepository) RemoveItem(itemID uint) error {
	return r.db.Delete(&models.CartItem{}, itemID).Error
}

func (r *cartRepository) GetCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	if err := r.db.Preload("Items").Preload("Items.Product").First(&cart, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) ClearCart(userID uint) error {
	var cart models.Cart
	if err := r.db.First(&cart, "user_id = ?", userID).Error; err != nil {
		return err
	}
	return r.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error
}

func (r *cartRepository) GetCartItems(userID uint) ([]models.CartItem, error) {
	var cart models.Cart
	if err := r.db.First(&cart, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	var items []models.CartItem
	if err := r.db.Preload("Product").Where("cart_id = ?", cart.ID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

