package repositories

import (
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ProductVariantRepository interface {
	// Variants
	CreateVariant(variant *models.ProductVariant) error
	FindVariantByID(id uint) (*models.ProductVariant, error)
	FindVariantsByProductID(productID uint) ([]models.ProductVariant, error)
	FindVariantsByProductIDAndValue(productID uint, value string) ([]models.ProductVariant, error)
	UpdateVariant(variant *models.ProductVariant) error
	DeleteVariant(id uint) error

	// Variant Values
	CreateVariantValue(value *models.ProductVariantValue) (*models.ProductVariantValue, error)
	FindVariantValueByID(id uint) (*models.ProductVariantValue, error)
	FindVariantValuesByVariantID(variantID uint) ([]models.ProductVariantValue, error)
	UpdateVariantValue(value *models.ProductVariantValue) error
	DeleteVariantValue(id uint) error
	DeleteVariantValuesByVariantID(variantID uint) error

	// Variant Combinations
	CreateVariantCombination(combination *models.ProductVariantCombination) error
	FindVariantCombinationByID(id uint) (*models.ProductVariantCombination, error)
	FindVariantCombinationsByProductID(productID uint, limit, offset int) ([]models.ProductVariantCombination, error)
	FindVariantCombinationBySKU(sku string) (*models.ProductVariantCombination, error)
	UpdateVariantCombination(combination *models.ProductVariantCombination) error
	DeleteVariantCombination(id uint) error
	UpdateVariantCombinationStock(id uint, quantity int) error
}

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) ProductVariantRepository {
	return &productVariantRepository{db: db}
}

// Variant methods

func (r *productVariantRepository) CreateVariant(variant *models.ProductVariant) error {
	return r.db.Create(variant).Error
}

func (r *productVariantRepository) FindVariantByID(id uint) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	if err := r.db.Preload("Product").First(&variant, id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productVariantRepository) FindVariantsByProductID(productID uint) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	if err := r.db.Where("product_id = ?", productID).Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

func (r *productVariantRepository) FindVariantsByProductIDAndValue(productID uint, value string) ([]models.ProductVariant, error) {
	var variants []models.ProductVariant
	if err := r.db.Where("product_id = ? AND name = ?", productID, value).Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

func (r *productVariantRepository) UpdateVariant(variant *models.ProductVariant) error {
	return r.db.Save(variant).Error
}

func (r *productVariantRepository) DeleteVariant(id uint) error {
	return r.db.Delete(&models.ProductVariant{}, id).Error
}

// Variant Value methods

func (r *productVariantRepository) CreateVariantValue(value *models.ProductVariantValue) (*models.ProductVariantValue, error) {
	if err := r.db.Create(value).Error; err != nil {
		return nil, err
	}
	return value, nil
}

func (r *productVariantRepository) FindVariantValueByID(id uint) (*models.ProductVariantValue, error) {
	var value models.ProductVariantValue
	if err := r.db.First(&value, id).Error; err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *productVariantRepository) FindVariantValuesByVariantID(variantID uint) ([]models.ProductVariantValue, error) {
	var values []models.ProductVariantValue
	if err := r.db.Where("variant_id = ?", variantID).Find(&values).Error; err != nil {
		return nil, err
	}
	return values, nil
}

func (r *productVariantRepository) UpdateVariantValue(value *models.ProductVariantValue) error {
	return r.db.Save(value).Error
}

func (r *productVariantRepository) DeleteVariantValue(id uint) error {
	return r.db.Delete(&models.ProductVariantValue{}, id).Error
}

func (r *productVariantRepository) DeleteVariantValuesByVariantID(variantID uint) error {
	return r.db.Where("variant_id = ?", variantID).Delete(&models.ProductVariantValue{}).Error
}

// Variant Combination methods

func (r *productVariantRepository) CreateVariantCombination(combination *models.ProductVariantCombination) error {
	return r.db.Create(combination).Error
}

func (r *productVariantRepository) FindVariantCombinationByID(id uint) (*models.ProductVariantCombination, error) {
	var combination models.ProductVariantCombination
	if err := r.db.Preload("Product").First(&combination, id).Error; err != nil {
		return nil, err
	}
	return &combination, nil
}

func (r *productVariantRepository) FindVariantCombinationsByProductID(productID uint, limit, offset int) ([]models.ProductVariantCombination, error) {
	var combinations []models.ProductVariantCombination
	if err := r.db.Where("product_id = ?", productID).Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&combinations).Error; err != nil {
		return nil, err
	}
	return combinations, nil
}

func (r *productVariantRepository) FindVariantCombinationBySKU(sku string) (*models.ProductVariantCombination, error) {
	var combination models.ProductVariantCombination
	if err := r.db.Where("sku = ?", sku).Preload("Product").First(&combination).Error; err != nil {
		return nil, err
	}
	return &combination, nil
}

func (r *productVariantRepository) UpdateVariantCombination(combination *models.ProductVariantCombination) error {
	return r.db.Save(combination).Error
}

func (r *productVariantRepository) DeleteVariantCombination(id uint) error {
	return r.db.Delete(&models.ProductVariantCombination{}, id).Error
}

func (r *productVariantRepository) UpdateVariantCombinationStock(id uint, quantity int) error {
	return r.db.Model(&models.ProductVariantCombination{ID: id}).Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
