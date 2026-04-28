package repositories

import (
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

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindBySKU(sku string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *productRepository) FindAll(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("category = ?", category).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) FindActive(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&models.Product{ID: id}).Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
