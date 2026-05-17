package repositories

import (
	"log"

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
	log.Printf("[productVariantRepository.CreateVariant] INFO: Creating variant - productID=%d, name=%s, type=%s", variant.ProductID, variant.Name, variant.Type)
	if err := r.db.Create(variant).Error; err != nil {
		log.Printf("[productVariantRepository.CreateVariant] ERROR: Failed to create variant - productID=%d: %v", variant.ProductID, err)
		return err
	}
	log.Printf("[productVariantRepository.CreateVariant] INFO: Variant created successfully - variantID=%d", variant.ID)
	return nil
}

func (r *productVariantRepository) FindVariantByID(id uint) (*models.ProductVariant, error) {
	log.Printf("[productVariantRepository.FindVariantByID] INFO: Retrieving variant - variantID=%d", id)
	var variant models.ProductVariant
	if err := r.db.Preload("Product").First(&variant, id).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantByID] ERROR: Failed to find variant - variantID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantByID] INFO: Variant found - variantID=%d, name=%s", id, variant.Name)
	return &variant, nil
}

func (r *productVariantRepository) FindVariantsByProductID(productID uint) ([]models.ProductVariant, error) {
	log.Printf("[productVariantRepository.FindVariantsByProductID] INFO: Retrieving variants - productID=%d", productID)
	var variants []models.ProductVariant
	if err := r.db.Where("product_id = ?", productID).Find(&variants).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantsByProductID] ERROR: Failed to retrieve variants - productID=%d: %v", productID, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantsByProductID] INFO: Variants retrieved - productID=%d, count=%d", productID, len(variants))
	return variants, nil
}

func (r *productVariantRepository) FindVariantsByProductIDAndValue(productID uint, value string) ([]models.ProductVariant, error) {
	log.Printf("[productVariantRepository.FindVariantsByProductIDAndValue] INFO: Retrieving variants - productID=%d, name=%s", productID, value)
	var variants []models.ProductVariant
	if err := r.db.Where("product_id = ? AND name = ?", productID, value).Find(&variants).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantsByProductIDAndValue] ERROR: Failed to retrieve variants - productID=%d, name=%s: %v", productID, value, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantsByProductIDAndValue] INFO: Variants retrieved - productID=%d, name=%s, count=%d", productID, value, len(variants))
	return variants, nil
}

func (r *productVariantRepository) UpdateVariant(variant *models.ProductVariant) error {
	log.Printf("[productVariantRepository.UpdateVariant] INFO: Updating variant - variantID=%d, name=%s", variant.ID, variant.Name)
	if err := r.db.Save(variant).Error; err != nil {
		log.Printf("[productVariantRepository.UpdateVariant] ERROR: Failed to update variant - variantID=%d: %v", variant.ID, err)
		return err
	}
	log.Printf("[productVariantRepository.UpdateVariant] INFO: Variant updated successfully - variantID=%d", variant.ID)
	return nil
}

func (r *productVariantRepository) DeleteVariant(id uint) error {
	log.Printf("[productVariantRepository.DeleteVariant] INFO: Deleting variant - variantID=%d", id)
	if err := r.db.Delete(&models.ProductVariant{}, id).Error; err != nil {
		log.Printf("[productVariantRepository.DeleteVariant] ERROR: Failed to delete variant - variantID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantRepository.DeleteVariant] INFO: Variant deleted successfully - variantID=%d", id)
	return nil
}

// Variant Value methods

func (r *productVariantRepository) CreateVariantValue(value *models.ProductVariantValue) (*models.ProductVariantValue, error) {
	log.Printf("[productVariantRepository.CreateVariantValue] INFO: Creating variant value - variantID=%d, value=%s", value.VariantID, value.Value)
	if err := r.db.Create(value).Error; err != nil {
		log.Printf("[productVariantRepository.CreateVariantValue] ERROR: Failed to create variant value - variantID=%d: %v", value.VariantID, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.CreateVariantValue] INFO: Variant value created successfully - valueID=%d", value.ID)
	return value, nil
}

func (r *productVariantRepository) FindVariantValueByID(id uint) (*models.ProductVariantValue, error) {
	log.Printf("[productVariantRepository.FindVariantValueByID] INFO: Retrieving variant value - valueID=%d", id)
	var value models.ProductVariantValue
	if err := r.db.First(&value, id).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantValueByID] ERROR: Failed to find variant value - valueID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantValueByID] INFO: Variant value found - valueID=%d, value=%s", id, value.Value)
	return &value, nil
}

func (r *productVariantRepository) FindVariantValuesByVariantID(variantID uint) ([]models.ProductVariantValue, error) {
	log.Printf("[productVariantRepository.FindVariantValuesByVariantID] INFO: Retrieving variant values - variantID=%d", variantID)
	var values []models.ProductVariantValue
	if err := r.db.Where("variant_id = ?", variantID).Find(&values).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantValuesByVariantID] ERROR: Failed to retrieve variant values - variantID=%d: %v", variantID, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantValuesByVariantID] INFO: Variant values retrieved - variantID=%d, count=%d", variantID, len(values))
	return values, nil
}

func (r *productVariantRepository) UpdateVariantValue(value *models.ProductVariantValue) error {
	log.Printf("[productVariantRepository.UpdateVariantValue] INFO: Updating variant value - valueID=%d, value=%s", value.ID, value.Value)
	if err := r.db.Save(value).Error; err != nil {
		log.Printf("[productVariantRepository.UpdateVariantValue] ERROR: Failed to update variant value - valueID=%d: %v", value.ID, err)
		return err
	}
	log.Printf("[productVariantRepository.UpdateVariantValue] INFO: Variant value updated successfully - valueID=%d", value.ID)
	return nil
}

func (r *productVariantRepository) DeleteVariantValue(id uint) error {
	log.Printf("[productVariantRepository.DeleteVariantValue] INFO: Deleting variant value - valueID=%d", id)
	if err := r.db.Delete(&models.ProductVariantValue{}, id).Error; err != nil {
		log.Printf("[productVariantRepository.DeleteVariantValue] ERROR: Failed to delete variant value - valueID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantRepository.DeleteVariantValue] INFO: Variant value deleted successfully - valueID=%d", id)
	return nil
}

func (r *productVariantRepository) DeleteVariantValuesByVariantID(variantID uint) error {
	log.Printf("[productVariantRepository.DeleteVariantValuesByVariantID] INFO: Deleting all values for variant - variantID=%d", variantID)
	if err := r.db.Where("variant_id = ?", variantID).Delete(&models.ProductVariantValue{}).Error; err != nil {
		log.Printf("[productVariantRepository.DeleteVariantValuesByVariantID] ERROR: Failed to delete values for variant - variantID=%d: %v", variantID, err)
		return err
	}
	log.Printf("[productVariantRepository.DeleteVariantValuesByVariantID] INFO: All values for variant deleted successfully - variantID=%d", variantID)
	return nil
}

// Variant Combination methods

func (r *productVariantRepository) CreateVariantCombination(combination *models.ProductVariantCombination) error {
	log.Printf("[productVariantRepository.CreateVariantCombination] INFO: Creating combination - productID=%d, SKU=%s, stock=%d", combination.ProductID, combination.SKU, combination.Stock)
	if err := r.db.Create(combination).Error; err != nil {
		log.Printf("[productVariantRepository.CreateVariantCombination] ERROR: Failed to create combination - SKU=%s: %v", combination.SKU, err)
		return err
	}
	log.Printf("[productVariantRepository.CreateVariantCombination] INFO: Combination created successfully - combinationID=%d, SKU=%s", combination.ID, combination.SKU)
	return nil
}

func (r *productVariantRepository) FindVariantCombinationByID(id uint) (*models.ProductVariantCombination, error) {
	log.Printf("[productVariantRepository.FindVariantCombinationByID] INFO: Retrieving combination - combinationID=%d", id)
	var combination models.ProductVariantCombination
	if err := r.db.Preload("Product").First(&combination, id).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantCombinationByID] ERROR: Failed to find combination - combinationID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantCombinationByID] INFO: Combination found - combinationID=%d, SKU=%s", id, combination.SKU)
	return &combination, nil
}

func (r *productVariantRepository) FindVariantCombinationsByProductID(productID uint, limit, offset int) ([]models.ProductVariantCombination, error) {
	log.Printf("[productVariantRepository.FindVariantCombinationsByProductID] INFO: Retrieving combinations - productID=%d, limit=%d, offset=%d", productID, limit, offset)
	var combinations []models.ProductVariantCombination
	if err := r.db.Where("product_id = ?", productID).Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&combinations).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantCombinationsByProductID] ERROR: Failed to retrieve combinations - productID=%d: %v", productID, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantCombinationsByProductID] INFO: Combinations retrieved - productID=%d, count=%d", productID, len(combinations))
	return combinations, nil
}

func (r *productVariantRepository) FindVariantCombinationBySKU(sku string) (*models.ProductVariantCombination, error) {
	log.Printf("[productVariantRepository.FindVariantCombinationBySKU] INFO: Retrieving combination - SKU=%s", sku)
	var combination models.ProductVariantCombination
	if err := r.db.Where("sku = ?", sku).Preload("Product").First(&combination).Error; err != nil {
		log.Printf("[productVariantRepository.FindVariantCombinationBySKU] ERROR: Failed to find combination - SKU=%s: %v", sku, err)
		return nil, err
	}
	log.Printf("[productVariantRepository.FindVariantCombinationBySKU] INFO: Combination found - combinationID=%d, SKU=%s", combination.ID, sku)
	return &combination, nil
}

func (r *productVariantRepository) UpdateVariantCombination(combination *models.ProductVariantCombination) error {
	log.Printf("[productVariantRepository.UpdateVariantCombination] INFO: Updating combination - combinationID=%d, SKU=%s", combination.ID, combination.SKU)
	if err := r.db.Save(combination).Error; err != nil {
		log.Printf("[productVariantRepository.UpdateVariantCombination] ERROR: Failed to update combination - combinationID=%d: %v", combination.ID, err)
		return err
	}
	log.Printf("[productVariantRepository.UpdateVariantCombination] INFO: Combination updated successfully - combinationID=%d", combination.ID)
	return nil
}

func (r *productVariantRepository) DeleteVariantCombination(id uint) error {
	log.Printf("[productVariantRepository.DeleteVariantCombination] INFO: Deleting combination - combinationID=%d", id)
	if err := r.db.Delete(&models.ProductVariantCombination{}, id).Error; err != nil {
		log.Printf("[productVariantRepository.DeleteVariantCombination] ERROR: Failed to delete combination - combinationID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantRepository.DeleteVariantCombination] INFO: Combination deleted successfully - combinationID=%d", id)
	return nil
}

func (r *productVariantRepository) UpdateVariantCombinationStock(id uint, quantity int) error {
	log.Printf("[productVariantRepository.UpdateVariantCombinationStock] INFO: Updating combination stock - combinationID=%d, quantityChange=%d", id, quantity)
	if err := r.db.Model(&models.ProductVariantCombination{ID: id}).Update("stock", gorm.Expr("stock + ?", quantity)).Error; err != nil {
		log.Printf("[productVariantRepository.UpdateVariantCombinationStock] ERROR: Failed to update combination stock - combinationID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantRepository.UpdateVariantCombinationStock] INFO: Combination stock updated successfully - combinationID=%d", id)
	return nil
}
