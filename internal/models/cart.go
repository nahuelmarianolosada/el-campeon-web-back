package models

import (
	"gorm.io/gorm"
	"time"
)

type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CartID    uint      `gorm:"not null" json:"cart_id"`
	Cart      *Cart     `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"-"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // Precio al momento de agregar al carrito
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Cart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Items     []CartItem     `gorm:"foreignKey:CartID" json:"items,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type CartResponse struct {
	ID     uint               `json:"id"`
	UserID uint               `json:"user_id"`
	Items  []CartItemResponse `json:"items"`
	Total  float64            `json:"total"`
}

type CartItemResponse struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	Subtotal  float64         `json:"subtotal"` // Quantity * Price
}
