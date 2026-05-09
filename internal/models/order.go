package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Order struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	OrderNumber     string            `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_number"`
	UserID          uint              `gorm:"not null" json:"user_id"`
	User            *User             `gorm:"foreignKey:UserID" json:"-"`
	Items           []OrderItem       `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Status          string            `gorm:"type:ENUM('PENDING','CONFIRMED','SHIPPED','DELIVERED','CANCELLED');default:'PENDING'" json:"status"`
	Subtotal        float64           `gorm:"not null" json:"subtotal"`
	Tax             float64           `gorm:"not null;default:0" json:"tax"`
	Total           float64           `gorm:"not null" json:"total"`
	ShippingAddress datatypes.JSONMap `gorm:"type:JSON" json:"shipping_address"`
	DeliveryMethod  string            `gorm:"type:ENUM('shipping','pickup-libreria','pickup-jugueteria');default:'shipping'" json:"delivery_method"`
	Notes           string            `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `gorm:"index" json:"-"`
}

type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // Precio al momento de la orden
	CreatedAt time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	ShippingAddress map[string]interface{} `json:"shipping_address" binding:"required"`
	DeliveryMethod  string                 `json:"delivery_method" binding:"required,oneof=shipping pickup-libreria pickup-jugueteria"`
	Notes           string                 `json:"notes"`
}

type OrderResponse struct {
	ID              uint                   `json:"id"`
	OrderNumber     string                 `json:"order_number"`
	UserID          uint                   `json:"user_id"`
	Items           []OrderItemResponse    `json:"items"`
	Status          string                 `json:"status"`
	Subtotal        float64                `json:"subtotal"`
	Tax             float64                `json:"tax"`
	Total           float64                `json:"total"`
	ShippingAddress map[string]interface{} `json:"shipping_address"`
	DeliveryMethod  string                 `json:"delivery_method"`
	Notes           string                 `json:"notes"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	Subtotal  float64         `json:"subtotal"` // Quantity * Price
}
