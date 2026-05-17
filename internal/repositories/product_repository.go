package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	FindByID(id uint) (*models.Product, error)
	FindBySKU(sku string) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id uint) error
	FindAll(limit, offset int) ([]models.Product, error)
	FindByCategory(category string, limit, offset int) ([]models.Product, error)
	FindActive(limit, offset int) ([]models.Product, error)
	UpdateStock(id uint, quantity int) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// orderImages ordena las imágenes precargadas por display_order.
func orderImages(db *gorm.DB) *gorm.DB {
	return db.Order("product_images.display_order ASC")
}

func (r *productRepository) Create(product *models.Product) error {
	log.Printf("[productRepository.Create] INFO: Creating product - SKU=%s, name=%s, retailPrice=%.2f, stock=%d", product.SKU, product.Name, product.PriceRetail, product.Stock)

	if err := r.db.Create(product).Error; err != nil {
		log.Printf("[productRepository.Create] ERROR: Failed to create product - SKU=%s: %v", product.SKU, err)
		return err
	}
	log.Printf("[productRepository.Create] INFO: Product created successfully - productID=%d, SKU=%s", product.ID, product.SKU)
	return nil
}

func (r *productRepository) FindByID(id uint) (*models.Product, error) {
	log.Printf("[productRepository.FindByID] INFO: Retrieving product - productID=%d", id)

	var product models.Product
	if err := r.db.Preload("Images", orderImages).First(&product, id).Error; err != nil {
		log.Printf("[productRepository.FindByID] ERROR: Failed to find product - productID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[productRepository.FindByID] INFO: Product found - productID=%d, SKU=%s, name=%s, stock=%d", product.ID, product.SKU, product.Name, product.Stock)
	return &product, nil
}

func (r *productRepository) FindBySKU(sku string) (*models.Product, error) {
	log.Printf("[productRepository.FindBySKU] INFO: Retrieving product - SKU=%s", sku)

	var product models.Product
	if err := r.db.Preload("Images", orderImages).Where("sku = ?", sku).First(&product).Error; err != nil {
		log.Printf("[productRepository.FindBySKU] ERROR: Failed to find product - SKU=%s: %v", sku, err)
		return nil, err
	}
	log.Printf("[productRepository.FindBySKU] INFO: Product found - productID=%d, SKU=%s, stock=%d", product.ID, sku, product.Stock)
	return &product, nil
}

func (r *productRepository) Update(product *models.Product) error {
	log.Printf("[productRepository.Update] INFO: Updating product - productID=%d, SKU=%s, name=%s", product.ID, product.SKU, product.Name)

	if err := r.db.Save(product).Error; err != nil {
		log.Printf("[productRepository.Update] ERROR: Failed to update product - productID=%d: %v", product.ID, err)
		return err
	}
	log.Printf("[productRepository.Update] INFO: Product updated successfully - productID=%d", product.ID)
	return nil
}

func (r *productRepository) Delete(id uint) error {
	log.Printf("[productRepository.Delete] INFO: Deleting product - productID=%d", id)

	if err := r.db.Delete(&models.Product{}, id).Error; err != nil {
		log.Printf("[productRepository.Delete] ERROR: Failed to delete product - productID=%d: %v", id, err)
		return err
	}
	log.Printf("[productRepository.Delete] INFO: Product deleted successfully - productID=%d", id)
	return nil
}

func (r *productRepository) FindAll(limit, offset int) ([]models.Product, error) {
	log.Printf("[productRepository.FindAll] INFO: Listing products - limit=%d, offset=%d", limit, offset)

	var products []models.Product
	if err := r.db.Preload("Images", orderImages).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		log.Printf("[productRepository.FindAll] ERROR: Failed to list products: %v", err)
		return nil, err
	}
	log.Printf("[productRepository.FindAll] INFO: Products listed - productCount=%d", len(products))
	return products, nil
}

func (r *productRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	log.Printf("[productRepository.FindByCategory] INFO: Listing products by category - category=%s, limit=%d, offset=%d", category, limit, offset)

	var products []models.Product
	if err := r.db.Preload("Images", orderImages).Where("category = ?", category).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		log.Printf("[productRepository.FindByCategory] ERROR: Failed to list products - category=%s: %v", category, err)
		return nil, err
	}
	log.Printf("[productRepository.FindByCategory] INFO: Products listed - category=%s, productCount=%d", category, len(products))
	return products, nil
}

func (r *productRepository) FindActive(limit, offset int) ([]models.Product, error) {
	log.Printf("[productRepository.FindActive] INFO: Listing active products - limit=%d, offset=%d", limit, offset)

	var products []models.Product
	if err := r.db.Preload("Images", orderImages).Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		log.Printf("[productRepository.FindActive] ERROR: Failed to list active products: %v", err)
		return nil, err
	}
	log.Printf("[productRepository.FindActive] INFO: Active products listed - productCount=%d", len(products))
	return products, nil
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	log.Printf("[productRepository.UpdateStock] INFO: Updating product stock - productID=%d, quantityChange=%d", id, quantity)

	if err := r.db.Model(&models.Product{ID: id}).Update("stock", gorm.Expr("stock + ?", quantity)).Error; err != nil {
		log.Printf("[productRepository.UpdateStock] ERROR: Failed to update stock - productID=%d: %v", id, err)
		return err
	}
	log.Printf("[productRepository.UpdateStock] INFO: Stock updated successfully - productID=%d, quantityChange=%d", id, quantity)
	return nil
}
