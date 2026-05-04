package variant

import (
	"encoding/json"
	"fmt"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type ProductVariantService interface {
	// Variants
	CreateVariant(productID uint, req *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error)
	GetVariant(id uint) (*models.ProductVariantResponse, error)
	GetProductVariants(productID uint) ([]models.ProductVariantResponse, error)
	UpdateVariant(id uint, req *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error)
	DeleteVariant(id uint) error

	// Variant Combinations
	CreateVariantCombination(productID uint, req *models.CreateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error)
	GetVariantCombination(id uint) (*models.ProductVariantCombinationResponse, error)
	GetProductVariantCombinations(productID uint, limit, offset int) ([]models.ProductVariantCombinationResponse, error)
	GetVariantCombinationBySKU(sku string) (*models.ProductVariantCombinationResponse, error)
	UpdateVariantCombination(id uint, req *models.UpdateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error)
	DeleteVariantCombination(id uint) error
}

type productVariantService struct {
	variantRepo repositories.ProductVariantRepository
	productRepo repositories.ProductRepository
}

func NewProductVariantService(
	variantRepo repositories.ProductVariantRepository,
	productRepo repositories.ProductRepository,
) ProductVariantService {
	return &productVariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
	}
}

// Variant methods

func (s *productVariantService) CreateVariant(productID uint, req *models.CreateProductVariantRequest) (*models.ProductVariantResponse, error) {
	// Verify product exists
	_, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	variant := &models.ProductVariant{
		ProductID: productID,
		Name:      req.Name,
		Type:      req.Type,
	}

	if err := s.variantRepo.CreateVariant(variant); err != nil {
		return nil, fmt.Errorf("error creating variant: %w", err)
	}

	// Create variant values
	for _, value := range req.Values {
		variantValue := &models.ProductVariantValue{
			VariantID: variant.ID,
			Value:     value,
		}
		if err := s.variantRepo.CreateVariantValue(variantValue); err != nil {
			return nil, fmt.Errorf("error creating variant value: %w", err)
		}
	}

	return s.toVariantResponse(variant, req.Values), nil
}

func (s *productVariantService) GetVariant(id uint) (*models.ProductVariantResponse, error) {
	variant, err := s.variantRepo.FindVariantByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding variant: %w", err)
	}

	values, err := s.variantRepo.FindVariantValuesByVariantID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding variant values: %w", err)
	}

	var valueStrings []string
	for _, v := range values {
		valueStrings = append(valueStrings, v.Value)
	}

	return s.toVariantResponse(variant, valueStrings), nil
}

func (s *productVariantService) GetProductVariants(productID uint) ([]models.ProductVariantResponse, error) {
	variants, err := s.variantRepo.FindVariantsByProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("error finding product variants: %w", err)
	}

	var responses []models.ProductVariantResponse
	for _, variant := range variants {
		values, err := s.variantRepo.FindVariantValuesByVariantID(variant.ID)
		if err != nil {
			return nil, fmt.Errorf("error finding variant values: %w", err)
		}

		var valueStrings []string
		for _, v := range values {
			valueStrings = append(valueStrings, v.Value)
		}

		responses = append(responses, *s.toVariantResponse(&variant, valueStrings))
	}

	return responses, nil
}

func (s *productVariantService) UpdateVariant(id uint, req *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error) {
	variant, err := s.variantRepo.FindVariantByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding variant: %w", err)
	}

	if req.Name != nil {
		variant.Name = *req.Name
	}

	if err := s.variantRepo.UpdateVariant(variant); err != nil {
		return nil, fmt.Errorf("error updating variant: %w", err)
	}

	// Update variant values if provided
	if len(req.Values) > 0 {
		// Delete existing values
		if err := s.variantRepo.DeleteVariantValuesByVariantID(id); err != nil {
			return nil, fmt.Errorf("error deleting variant values: %w", err)
		}

		// Create new values
		for _, value := range req.Values {
			variantValue := &models.ProductVariantValue{
				VariantID: id,
				Value:     value,
			}
			if err := s.variantRepo.CreateVariantValue(variantValue); err != nil {
				return nil, fmt.Errorf("error creating variant value: %w", err)
			}
		}
	}

	return s.GetVariant(id)
}

func (s *productVariantService) DeleteVariant(id uint) error {
	// Delete variant values first
	if err := s.variantRepo.DeleteVariantValuesByVariantID(id); err != nil {
		return fmt.Errorf("error deleting variant values: %w", err)
	}

	return s.variantRepo.DeleteVariant(id)
}

// Variant Combination methods

func (s *productVariantService) CreateVariantCombination(productID uint, req *models.CreateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error) {
	// Verify product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Convert variant combination map to JSON
	combinationJSON, err := json.Marshal(req.VariantCombination)
	if err != nil {
		return nil, fmt.Errorf("error marshaling variant combination: %w", err)
	}

	combination := &models.ProductVariantCombination{
		ProductID:          productID,
		SKU:                req.SKU,
		VariantCombination: string(combinationJSON),
		Stock:              req.Stock,
		PriceAdjustment:    req.PriceAdjustment,
		ImageURL:           req.ImageURL,
		IsActive:           true,
	}

	if err := s.variantRepo.CreateVariantCombination(combination); err != nil {
		return nil, fmt.Errorf("error creating variant combination: %w", err)
	}

	return s.toVariantCombinationResponse(combination, product), nil
}

func (s *productVariantService) GetVariantCombination(id uint) (*models.ProductVariantCombinationResponse, error) {
	combination, err := s.variantRepo.FindVariantCombinationByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding variant combination: %w", err)
	}

	product, err := s.productRepo.FindByID(combination.ProductID)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}

	return s.toVariantCombinationResponse(combination, product), nil
}

func (s *productVariantService) GetProductVariantCombinations(productID uint, limit, offset int) ([]models.ProductVariantCombinationResponse, error) {
	// Verify product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	combinations, err := s.variantRepo.FindVariantCombinationsByProductID(productID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding variant combinations: %w", err)
	}

	var responses []models.ProductVariantCombinationResponse
	for _, combination := range combinations {
		responses = append(responses, *s.toVariantCombinationResponse(&combination, product))
	}

	return responses, nil
}

func (s *productVariantService) GetVariantCombinationBySKU(sku string) (*models.ProductVariantCombinationResponse, error) {
	combination, err := s.variantRepo.FindVariantCombinationBySKU(sku)
	if err != nil {
		return nil, fmt.Errorf("error finding variant combination: %w", err)
	}

	return s.toVariantCombinationResponse(combination, combination.Product), nil
}

func (s *productVariantService) UpdateVariantCombination(id uint, req *models.UpdateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error) {
	combination, err := s.variantRepo.FindVariantCombinationByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding variant combination: %w", err)
	}

	if req.Stock != nil {
		combination.Stock = *req.Stock
	}
	if req.PriceAdjustment != nil {
		combination.PriceAdjustment = *req.PriceAdjustment
	}
	if req.ImageURL != nil {
		combination.ImageURL = *req.ImageURL
	}
	if req.IsActive != nil {
		combination.IsActive = *req.IsActive
	}

	if err := s.variantRepo.UpdateVariantCombination(combination); err != nil {
		return nil, fmt.Errorf("error updating variant combination: %w", err)
	}

	return s.GetVariantCombination(id)
}

func (s *productVariantService) DeleteVariantCombination(id uint) error {
	return s.variantRepo.DeleteVariantCombination(id)
}

// Helper functions

func (s *productVariantService) toVariantResponse(variant *models.ProductVariant, values []string) *models.ProductVariantResponse {
	valueResponses := make([]models.ProductVariantValueResponse, len(values))
	for i, v := range values {
		valueResponses[i] = models.ProductVariantValueResponse{
			Value: v,
		}
	}

	return &models.ProductVariantResponse{
		ID:     variant.ID,
		Name:   variant.Name,
		Type:   variant.Type,
		Values: valueResponses,
	}
}

func (s *productVariantService) toVariantCombinationResponse(combination *models.ProductVariantCombination, product *models.Product) *models.ProductVariantCombinationResponse {
	var variantMap map[string]string
	if err := json.Unmarshal([]byte(combination.VariantCombination), &variantMap); err != nil {
		variantMap = make(map[string]string)
	}

	finalPrice := product.PriceRetail + combination.PriceAdjustment

	return &models.ProductVariantCombinationResponse{
		ID:                 combination.ID,
		SKU:                combination.SKU,
		VariantCombination: variantMap,
		Stock:              combination.Stock,
		PriceAdjustment:    combination.PriceAdjustment,
		ImageURL:           combination.ImageURL,
		FinalPrice:         finalPrice,
		IsActive:           combination.IsActive,
		CreatedAt:          combination.CreatedAt,
	}
}


