package product

import (
	"fmt"
	"io"
	"log"
	"strings"

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
	ImportProducts(file io.Reader, fileName string, dryRun bool) (*ImportResult, error)
}

type productService struct {
	productRepo repositories.ProductRepository
	imageRepo   repositories.ProductImageRepository
}

func NewProductService(productRepo repositories.ProductRepository, imageRepo repositories.ProductImageRepository) ProductService {
	return &productService{
		productRepo: productRepo,
		imageRepo:   imageRepo,
	}
}

// buildProductImages convierte una lista de URLs en imágenes ordenadas,
// descartando las entradas vacías.
func buildProductImages(imgs []models.ProductImage) []models.ProductImage {
	images := make([]models.ProductImage, 0, len(imgs))
	for _, img := range imgs {
		trimmed := strings.TrimSpace(img.ImageURL)
		if trimmed == "" {
			continue
		}
		images = append(images, models.ProductImage{
			ImageURL:     trimmed,
			DisplayOrder: len(images),
		})
	}
	return images
}

func (s *productService) CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error) {
	log.Printf("[productService.CreateProduct] INFO: Starting product creation - SKU=%s, name=%s, retailPrice=%.2f, wholesalePrice=%.2f", req.SKU, req.Name, req.PriceRetail, req.PriceWholesale)

	product := &models.Product{
		SKU:             req.SKU,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		PriceRetail:     req.PriceRetail,
		PriceWholesale:  req.PriceWholesale,
		Stock:           req.Stock,
		MinBulkQuantity: req.MinBulkQuantity,
		Images:          buildProductImages(req.ImageURLs),
		IsActive:        true,
	}

	if err := s.productRepo.Create(product); err != nil {
		log.Printf("[productService.CreateProduct] ERROR: Failed to create product - SKU=%s: %v", req.SKU, err)
		return nil, fmt.Errorf("error creating product: %w", err)
	}
	log.Printf("[productService.CreateProduct] INFO: Product created successfully - productID=%d, SKU=%s", product.ID, req.SKU)

	return s.toProductResponse(product), nil
}

func (s *productService) GetProductByID(id uint) (*models.ProductResponse, error) {
	log.Printf("[productService.GetProductByID] INFO: Retrieving product - productID=%d", id)

	product, err := s.productRepo.FindByID(id)
	if err != nil {
		log.Printf("[productService.GetProductByID] ERROR: Failed to find product - productID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding product: %w", err)
	}
	log.Printf("[productService.GetProductByID] INFO: Product found - productID=%d, SKU=%s, name=%s, stock=%d", product.ID, product.SKU, product.Name, product.Stock)

	return s.toProductResponse(product), nil
}

func (s *productService) GetProductBySKU(sku string) (*models.ProductResponse, error) {
	log.Printf("[productService.GetProductBySKU] INFO: Retrieving product - SKU=%s", sku)

	product, err := s.productRepo.FindBySKU(sku)
	if err != nil {
		log.Printf("[productService.GetProductBySKU] ERROR: Failed to find product - SKU=%s: %v", sku, err)
		return nil, fmt.Errorf("error finding product: %w", err)
	}
	log.Printf("[productService.GetProductBySKU] INFO: Product found - productID=%d, SKU=%s, stock=%d", product.ID, sku, product.Stock)

	return s.toProductResponse(product), nil
}

func (s *productService) UpdateProduct(id uint, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	log.Printf("[productService.UpdateProduct] INFO: Starting product update - productID=%d", id)

	product, err := s.productRepo.FindByID(id)
	if err != nil {
		log.Printf("[productService.UpdateProduct] ERROR: Failed to find product - productID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding product: %w", err)
	}

	// Actualizar campos si están presentes
	if req.Name != nil {
		log.Printf("[productService.UpdateProduct] INFO: Updating name - productID=%d, oldName=%s, newName=%s", id, product.Name, *req.Name)
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.PriceRetail != nil {
		log.Printf("[productService.UpdateProduct] INFO: Updating retail price - productID=%d, oldPrice=%.2f, newPrice=%.2f", id, product.PriceRetail, *req.PriceRetail)
		product.PriceRetail = *req.PriceRetail
	}
	if req.PriceWholesale != nil {
		log.Printf("[productService.UpdateProduct] INFO: Updating wholesale price - productID=%d, oldPrice=%.2f, newPrice=%.2f", id, product.PriceWholesale, *req.PriceWholesale)
		product.PriceWholesale = *req.PriceWholesale
	}
	if req.Stock != nil {
		log.Printf("[productService.UpdateProduct] INFO: Updating stock - productID=%d, oldStock=%d, newStock=%d", id, product.Stock, *req.Stock)
		product.Stock = *req.Stock
	}
	if req.MinBulkQuantity != nil {
		product.MinBulkQuantity = *req.MinBulkQuantity
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	existingImages := product.Images
	// Evitar que Save vuelva a tocar las asociaciones; las imágenes se gestionan aparte.
	product.Images = nil

	if err := s.productRepo.Update(product); err != nil {
		log.Printf("[productService.UpdateProduct] ERROR: Failed to update product - productID=%d: %v", id, err)
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	if req.ImageURLs != nil {
		urls := make([]models.ProductImage, len(*req.ImageURLs))
		copy(urls, *req.ImageURLs)
		if err := s.imageRepo.ReplaceForProduct(id, urls); err != nil {
			log.Printf("[productService.UpdateProduct] ERROR: Failed to update images - productID=%d: %v", id, err)
			return nil, fmt.Errorf("error updating product images: %w", err)
		}
		images, err := s.imageRepo.FindByProductID(id)
		if err != nil {
			log.Printf("[productService.UpdateProduct] ERROR: Failed to reload images - productID=%d: %v", id, err)
			return nil, fmt.Errorf("error loading product images: %w", err)
		}
		product.Images = images
	} else {
		product.Images = existingImages
	}
	log.Printf("[productService.UpdateProduct] INFO: Product updated successfully - productID=%d", id)

	return s.toProductResponse(product), nil
}

func (s *productService) DeleteProduct(id uint) error {
	log.Printf("[productService.DeleteProduct] INFO: Deleting product - productID=%d", id)

	if err := s.productRepo.Delete(id); err != nil {
		log.Printf("[productService.DeleteProduct] ERROR: Failed to delete product - productID=%d: %v", id, err)
		return err
	}
	log.Printf("[productService.DeleteProduct] INFO: Product deleted successfully - productID=%d", id)
	return nil
}

func (s *productService) ListProducts(limit, offset int) ([]models.ProductResponse, error) {
	log.Printf("[productService.ListProducts] INFO: Listing products - limit=%d, offset=%d", limit, offset)

	products, err := s.productRepo.FindAll(limit, offset)
	if err != nil {
		log.Printf("[productService.ListProducts] ERROR: Failed to list products: %v", err)
		return nil, fmt.Errorf("error listing products: %w", err)
	}
	log.Printf("[productService.ListProducts] INFO: Products listed - productCount=%d", len(products))

	return s.toProductResponses(products), nil
}

func (s *productService) ListProductsByCategory(category string, limit, offset int) ([]models.ProductResponse, error) {
	log.Printf("[productService.ListProductsByCategory] INFO: Listing products by category - category=%s, limit=%d, offset=%d", category, limit, offset)

	products, err := s.productRepo.FindByCategory(category, limit, offset)
	if err != nil {
		log.Printf("[productService.ListProductsByCategory] ERROR: Failed to list products - category=%s: %v", category, err)
		return nil, fmt.Errorf("error listing products by category: %w", err)
	}
	log.Printf("[productService.ListProductsByCategory] INFO: Products listed - category=%s, productCount=%d", category, len(products))

	return s.toProductResponses(products), nil
}

func (s *productService) ListActiveProducts(limit, offset int) ([]models.ProductResponse, error) {
	log.Printf("[productService.ListActiveProducts] INFO: Listing active products - limit=%d, offset=%d", limit, offset)

	products, err := s.productRepo.FindActive(limit, offset)
	if err != nil {
		log.Printf("[productService.ListActiveProducts] ERROR: Failed to list active products: %v", err)
		return nil, fmt.Errorf("error listing active products: %w", err)
	}
	log.Printf("[productService.ListActiveProducts] INFO: Active products listed - productCount=%d", len(products))

	return s.toProductResponses(products), nil
}

// GetPrice determina el precio basado en si es comprador mayorista y cantidad
func (s *productService) GetPrice(productID uint, isBulkBuyer bool, quantity int) (float64, error) {
	log.Printf("[productService.GetPrice] INFO: Calculating price - productID=%d, isBulkBuyer=%v, quantity=%d", productID, isBulkBuyer, quantity)

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		log.Printf("[productService.GetPrice] ERROR: Failed to find product - productID=%d: %v", productID, err)
		return 0, fmt.Errorf("error finding product: %w", err)
	}

	// Si es comprador mayorista y la cantidad excede el mínimo, aplicar precio mayorista
	if isBulkBuyer && quantity >= product.MinBulkQuantity {
		log.Printf("[productService.GetPrice] INFO: Bulk price applied - productID=%d, price=%.2f", productID, product.PriceWholesale)
		return product.PriceWholesale, nil
	}

	log.Printf("[productService.GetPrice] INFO: Retail price applied - productID=%d, price=%.2f", productID, product.PriceRetail)
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
		ImageURL:        models.PrimaryImageURL(product.Images),
		Images:          models.ToProductImageResponses(product.Images),
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
