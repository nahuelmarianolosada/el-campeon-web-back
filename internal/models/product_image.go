package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductImage struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	ImageURL     string         `gorm:"type:varchar(500);not null" json:"image_url"`
	DisplayOrder int            `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductImageResponse struct {
	ID           uint   `json:"id"`
	ImageURL     string `json:"image_url"`
	DisplayOrder int    `json:"display_order"`
}

// ToProductImageResponses convierte las imágenes de un producto a su DTO de respuesta.
func ToProductImageResponses(images []ProductImage) []ProductImageResponse {
	responses := make([]ProductImageResponse, 0, len(images))
	for _, img := range images {
		responses = append(responses, ProductImageResponse{
			ID:           img.ID,
			ImageURL:     img.ImageURL,
			DisplayOrder: img.DisplayOrder,
		})
	}
	return responses
}

// PrimaryImageURL devuelve la URL de la primera imagen (ordenadas por display_order),
// usada como miniatura en contextos que solo necesitan una imagen (carrito, órdenes).
func PrimaryImageURL(images []ProductImage) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].ImageURL
}
