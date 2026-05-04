package models

import (
	"time"

	"gorm.io/gorm"
)

// ProductVariant representa un tipo de variante de producto (ej: Color, Tamaño, Material)
type ProductVariant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	Product   *Product       `gorm:"foreignKey:ProductID" json:"-"`
	Name      string         `gorm:"not null" json:"name"` // ej: "Color", "Tamaño", "Material"
	Type      string         `gorm:"not null" json:"type"` // ej: "color", "size", "material"
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProductVariantValue representa un valor específico de una variante (ej: "Rojo" para Color)
type ProductVariantValue struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	VariantID uint           `gorm:"not null;index" json:"variant_id"`
	Variant   *ProductVariant `gorm:"foreignKey:VariantID" json:"-"`
	Value     string         `gorm:"not null" json:"value"` // ej: "Rojo", "Azul", "Grande", "Pequeño"
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProductVariantCombination representa una combinación específica de variantes con su propio SKU y stock
type ProductVariantCombination struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	ProductID             uint      `gorm:"not null;index" json:"product_id"`
	Product               *Product  `gorm:"foreignKey:ProductID" json:"-"`
	SKU                   string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"sku"` // ej: "PEN-RED-THIN"
	VariantCombination    string    `gorm:"not null" json:"variant_combination"`               // JSON serializado con las combinaciones
	Stock                 int       `gorm:"not null;default:0" json:"stock"`
	PriceAdjustment       float64   `gorm:"default:0" json:"price_adjustment"`       // Ajuste de precio adicional
	ImageURL              string    `json:"image_url"`
	IsActive              bool      `gorm:"default:true" json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// Request models for variants

type CreateProductVariantRequest struct {
	Name string `json:"name" binding:"required"`      // ej: "Color"
	Type string `json:"type" binding:"required"`      // ej: "color"
	Values []string `json:"values" binding:"required"` // ej: ["Rojo", "Azul", "Verde"]
}

type UpdateProductVariantRequest struct {
	Name   *string   `json:"name"`
	Values []string  `json:"values"` // Nueva lista de valores
}

type CreateProductVariantCombinationRequest struct {
	SKU                string             `json:"sku" binding:"required"`
	VariantCombination map[string]string `json:"variant_combination" binding:"required"` // ej: {"Color": "Rojo", "Tamaño": "Grande"}
	Stock              int                `json:"stock" binding:"gte=0"`
	PriceAdjustment    float64            `json:"price_adjustment"`
	ImageURL           string             `json:"image_url"`
}

type UpdateProductVariantCombinationRequest struct {
	Stock            *int    `json:"stock"`
	PriceAdjustment  *float64 `json:"price_adjustment"`
	ImageURL         *string `json:"image_url"`
	IsActive         *bool   `json:"is_active"`
}

// Response models

type ProductVariantResponse struct {
	ID     uint                         `json:"id"`
	Name   string                       `json:"name"`
	Type   string                       `json:"type"`
	Values []ProductVariantValueResponse `json:"values"`
}

type ProductVariantValueResponse struct {
	ID    uint   `json:"id"`
	Value string `json:"value"`
}

type ProductVariantCombinationResponse struct {
	ID                 uint              `json:"id"`
	SKU                string            `json:"sku"`
	VariantCombination map[string]string `json:"variant_combination"`
	Stock              int               `json:"stock"`
	PriceAdjustment    float64           `json:"price_adjustment"`
	ImageURL           string            `json:"image_url"`
	FinalPrice         float64           `json:"final_price"` // precio base del producto + ajuste
	IsActive           bool              `json:"is_active"`
	CreatedAt          time.Time         `json:"created_at"`
}

