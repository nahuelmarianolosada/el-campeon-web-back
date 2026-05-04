package product

import (
	"fmt"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type ProductService interface {
	CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error)
	GetProductByID(id uint) (*models.ProductResponse, error)
	GetProductBySKU(sku string) (*models.ProductResponse, error)
	UpdateProduct(id uint, req *models.UpdateProductRequest) (*models.ProductResponse, error)
	DeleteProduct(id uint) error
	ListProducts(limit, offset int) ([]models.ProductResponse, error)
	ListProductsByCategory(category string, limit, offset int) ([]models.ProductResponse, error)
	ListActiveProducts(limit, offset int) ([]models.ProductResponse, error)
	GetPrice(productID uint, isBulkBuyer bool, quantity int) (float64, error)
}

type productService struct {
	productRepo repositories.ProductRepository
}

func NewProductService(productRepo repositories.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

func (s *productService) CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error) {
	product := &models.Product{
		SKU:             req.SKU,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		PriceRetail:     req.PriceRetail,
		PriceWholesale:  req.PriceWholesale,
		Stock:           req.Stock,
		MinBulkQuantity: req.MinBulkQuantity,
		ImageURL:        req.ImageURL,
		IsActive:        true,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	return s.toProductResponse(product), nil
}

func (s *productService) GetProductByID(id uint) (*models.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}

	return s.toProductResponse(product), nil
}

func (s *productService) GetProductBySKU(sku string) (*models.ProductResponse, error) {
	product, err := s.productRepo.FindBySKU(sku)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}

	return s.toProductResponse(product), nil
}

func (s *productService) UpdateProduct(id uint, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}

	// Actualizar campos si están presentes
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.PriceRetail != nil {
		product.PriceRetail = *req.PriceRetail
	}
	if req.PriceWholesale != nil {
		product.PriceWholesale = *req.PriceWholesale
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.MinBulkQuantity != nil {
		product.MinBulkQuantity = *req.MinBulkQuantity
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	return s.toProductResponse(product), nil
}

func (s *productService) DeleteProduct(id uint) error {
	return s.productRepo.Delete(id)
}

func (s *productService) ListProducts(limit, offset int) ([]models.ProductResponse, error) {
	products, err := s.productRepo.FindAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}

	return s.toProductResponses(products), nil
}

func (s *productService) ListProductsByCategory(category string, limit, offset int) ([]models.ProductResponse, error) {
	products, err := s.productRepo.FindByCategory(category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing products by category: %w", err)
	}

	return s.toProductResponses(products), nil
}

func (s *productService) ListActiveProducts(limit, offset int) ([]models.ProductResponse, error) {
	products, err := s.productRepo.FindActive(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing active products: %w", err)
	}

	return s.toProductResponses(products), nil
}

// GetPrice determina el precio basado en si es comprador mayorista y cantidad
func (s *productService) GetPrice(productID uint, isBulkBuyer bool, quantity int) (float64, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return 0, fmt.Errorf("error finding product: %w", err)
	}

	// Si es comprador mayorista y la cantidad excede el mínimo, aplicar precio mayorista
	if isBulkBuyer && quantity >= product.MinBulkQuantity {
		return product.PriceWholesale, nil
	}

	return product.PriceRetail, nil
}

// Helper functions

func (s *productService) toProductResponse(product *models.Product) *models.ProductResponse {
	return &models.ProductResponse{
		ID:              product.ID,
		SKU:             product.SKU,
		Name:            product.Name,
		Description:     product.Description,
		Category:        product.Category,
		PriceRetail:     product.PriceRetail,
		PriceWholesale:  product.PriceWholesale,
		Stock:           product.Stock,
		MinBulkQuantity: product.MinBulkQuantity,
		ImageURL:        product.ImageURL,
		IsActive:        product.IsActive,
		HasVariants:     product.HasVariants,
		CreatedAt:       product.CreatedAt,
	}
}

func (s *productService) toProductResponses(products []models.Product) []models.ProductResponse {
	var responses []models.ProductResponse
	for _, product := range products {
		responses = append(responses, *s.toProductResponse(&product))
	}
	return responses
}
