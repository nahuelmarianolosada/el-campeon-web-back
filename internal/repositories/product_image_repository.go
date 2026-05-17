package repositories

import (
	"log"
	"strings"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ProductImageRepository interface {
	FindByProductID(productID uint) ([]models.ProductImage, error)
	ReplaceForProduct(productID uint, urls []models.ProductImage) error
}

type productImageRepository struct {
	db *gorm.DB
}

func NewProductImageRepository(db *gorm.DB) ProductImageRepository {
	return &productImageRepository{db: db}
}

func (r *productImageRepository) FindByProductID(productID uint) ([]models.ProductImage, error) {
	var images []models.ProductImage
	if err := r.db.Where("product_id = ?", productID).Order("display_order ASC").Find(&images).Error; err != nil {
		log.Printf("[productImageRepository.FindByProductID] ERROR: Failed to list images - productID=%d: %v", productID, err)
		return nil, err
	}
	return images, nil
}

// ReplaceForProduct reemplaza por completo el set de imágenes de un producto.
// El orden de la slice define el display_order de cada imagen.
func (r *productImageRepository) ReplaceForProduct(productID uint, imgs []models.ProductImage) error {
	log.Printf("[productImageRepository.ReplaceForProduct] INFO: Replacing images - productID=%d, count=%d", productID, len(imgs))

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&models.ProductImage{}).Error; err != nil {
			log.Printf("[productImageRepository.ReplaceForProduct] ERROR: Failed to delete old images - productID=%d: %v", productID, err)
			return err
		}

		images := make([]models.ProductImage, 0, len(imgs))
		for _, img := range imgs {
			trimmed := strings.TrimSpace(img.ImageURL)
			if trimmed == "" {
				continue
			}
			images = append(images, models.ProductImage{
				ProductID:    productID,
				ImageURL:     trimmed,
				DisplayOrder: len(images),
			})
		}

		if len(images) == 0 {
			return nil
		}

		if err := tx.Create(&images).Error; err != nil {
			log.Printf("[productImageRepository.ReplaceForProduct] ERROR: Failed to create images - productID=%d: %v", productID, err)
			return err
		}
		return nil
	})
}
