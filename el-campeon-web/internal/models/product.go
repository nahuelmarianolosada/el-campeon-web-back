package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	SKU             string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"sku"`
	Name            string         `gorm:"not null" json:"name"`
	Description     string         `gorm:"type:varchar(255)" json:"description"`
	Category        string         `gorm:"not null" json:"category"`
	PriceRetail     float64        `gorm:"not null" json:"price_retail"`    // Precio minorista
	PriceWholesale  float64        `gorm:"not null" json:"price_wholesale"` // Precio mayorista
	Stock           int            `gorm:"not null;default:0" json:"stock"`
	MinBulkQuantity int            `gorm:"default:10" json:"min_bulk_quantity"` // Cantidad mínima para aplicar mayorista
	ImageURL        string         `json:"image_url"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateProductRequest struct {
	SKU             string  `json:"sku" binding:"required,max=50"`
	Name            string  `json:"name" binding:"required,max=255"`
	Description     string  `json:"description"`
	Category        string  `json:"category" binding:"required"`
	PriceRetail     float64 `json:"price_retail" binding:"required,gt=0"`
	PriceWholesale  float64 `json:"price_wholesale" binding:"required,gt=0"`
	Stock           int     `json:"stock" binding:"required,gte=0"`
	MinBulkQuantity int     `json:"min_bulk_quantity"`
	ImageURL        string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name            *string  `json:"name"`
	Description     *string  `json:"description"`
	Category        *string  `json:"category"`
	PriceRetail     *float64 `json:"price_retail"`
	PriceWholesale  *float64 `json:"price_wholesale"`
	Stock           *int     `json:"stock"`
	MinBulkQuantity *int     `json:"min_bulk_quantity"`
	ImageURL        *string  `json:"image_url"`
	IsActive        *bool    `json:"is_active"`
}

type ProductResponse struct {
	ID              uint      `json:"id"`
	SKU             string    `json:"sku"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	PriceRetail     float64   `json:"price_retail"`
	PriceWholesale  float64   `json:"price_wholesale"`
	Stock           int       `json:"stock"`
	MinBulkQuantity int       `json:"min_bulk_quantity"`
	ImageURL        string    `json:"image_url"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}
