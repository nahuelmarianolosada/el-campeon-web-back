package variant

import (
	"encoding/json"
	"fmt"
	"log"

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
	log.Printf("[productVariantService.CreateVariant] INFO: Creating variant - productID=%d, name=%s, type=%s", productID, req.Name, req.Type)
	// Verify product exists
	_, err := s.productRepo.FindByID(productID)
	if err != nil {
		log.Printf("[productVariantService.CreateVariant] ERROR: Product not found - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if variant already exists
	var productVariants []models.ProductVariant
	if productVariants, err = s.variantRepo.FindVariantsByProductIDAndValue(productID, req.Name); err != nil {
		log.Printf("[productVariantService.CreateVariant] ERROR: Failed to search for existing variants - productID=%d, name=%s: %v", productID, req.Name, err)
		return nil, fmt.Errorf("error finding variants: %w", err)
	}

	if len(productVariants) > 0 {
		log.Printf("[productVariantService.CreateVariant] INFO: Variant already exists, updating values - productID=%d, variantID=%d", productID, productVariants[0].ID)
		variant := &productVariants[0]
		existingValues, err := s.variantRepo.FindVariantValuesByVariantID(variant.ID)
		if err != nil {
			log.Printf("[productVariantService.CreateVariant] ERROR: Failed to find existing values - variantID=%d: %v", variant.ID, err)
			return nil, fmt.Errorf("error finding variant values: %w", err)
		}

		existingValuesMap := make(map[string]bool)
		for _, v := range existingValues {
			existingValuesMap[v.Value] = true
		}

		for _, reqValue := range req.Values {
			if !existingValuesMap[reqValue] {
				newVariantValue := &models.ProductVariantValue{
					VariantID: variant.ID,
					Value:     reqValue,
				}
				createdValue, err := s.variantRepo.CreateVariantValue(newVariantValue)
				if err != nil {
					return nil, fmt.Errorf("error creating variant value: %w", err)
				}
				existingValues = append(existingValues, *createdValue)
				existingValuesMap[reqValue] = true
			}
		}
		return s.toVariantResponse(variant, existingValues), nil
	}

	variant := &models.ProductVariant{
		ProductID: productID,
		Name:      req.Name,
		Type:      req.Type,
	}

	if err := s.variantRepo.CreateVariant(variant); err != nil {
		log.Printf("[productVariantService.CreateVariant] ERROR: Failed to create variant - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("error creating variant: %w", err)
	}
	log.Printf("[productVariantService.CreateVariant] INFO: Variant created successfully - variantID=%d", variant.ID)

	// Create variant values
	var variantValues []models.ProductVariantValue
	for _, value := range req.Values {
		variantValue := &models.ProductVariantValue{
			VariantID: variant.ID,
			Value:     value,
		}
		if newVariantValue, err := s.variantRepo.CreateVariantValue(variantValue); err != nil {
			return nil, fmt.Errorf("error creating variant value: %w", err)
		} else {
			variantValues = append(variantValues, *newVariantValue)
		}
	}

	return s.toVariantResponse(variant, variantValues), nil
}

func (s *productVariantService) GetVariant(id uint) (*models.ProductVariantResponse, error) {
	log.Printf("[productVariantService.GetVariant] INFO: Retrieving variant - variantID=%d", id)
	variant, err := s.variantRepo.FindVariantByID(id)
	if err != nil {
		log.Printf("[productVariantService.GetVariant] ERROR: Failed to find variant - variantID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding variant: %w", err)
	}

	values, err := s.variantRepo.FindVariantValuesByVariantID(id)
	if err != nil {
		log.Printf("[productVariantService.GetVariant] ERROR: Failed to find variant values - variantID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding variant values: %w", err)
	}
	log.Printf("[productVariantService.GetVariant] INFO: Variant retrieved successfully - variantID=%d, valueCount=%d", id, len(values))

	return s.toVariantResponse(variant, values), nil
}

func (s *productVariantService) GetProductVariants(productID uint) ([]models.ProductVariantResponse, error) {
	log.Printf("[productVariantService.GetProductVariants] INFO: Retrieving variants for product - productID=%d", productID)
	variants, err := s.variantRepo.FindVariantsByProductID(productID)
	if err != nil {
		log.Printf("[productVariantService.GetProductVariants] ERROR: Failed to find product variants - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("error finding product variants: %w", err)
	}

	var responses []models.ProductVariantResponse
	for _, variant := range variants {
		values, err := s.variantRepo.FindVariantValuesByVariantID(variant.ID)
		if err != nil {
			return nil, fmt.Errorf("error finding variant values: %w", err)
		}

		responses = append(responses, *s.toVariantResponse(&variant, values))
	}

	return responses, nil
}

func (s *productVariantService) UpdateVariant(id uint, req *models.UpdateProductVariantRequest) (*models.ProductVariantResponse, error) {
	log.Printf("[productVariantService.UpdateVariant] INFO: Updating variant - variantID=%d", id)
	variant, err := s.variantRepo.FindVariantByID(id)
	if err != nil {
		log.Printf("[productVariantService.UpdateVariant] ERROR: Failed to find variant - variantID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding variant: %w", err)
	}

	if req.Name != nil {
		variant.Name = *req.Name
	}

	if err := s.variantRepo.UpdateVariant(variant); err != nil {
		log.Printf("[productVariantService.UpdateVariant] ERROR: Failed to update variant - variantID=%d: %v", id, err)
		return nil, fmt.Errorf("error updating variant: %w", err)
	}
	log.Printf("[productVariantService.UpdateVariant] INFO: Variant updated successfully - variantID=%d", id)

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
			if _, err := s.variantRepo.CreateVariantValue(variantValue); err != nil {
				return nil, fmt.Errorf("error creating variant value: %w", err)
			}
		}
	}

	return s.GetVariant(id)
}

func (s *productVariantService) DeleteVariant(id uint) error {
	log.Printf("[productVariantService.DeleteVariant] INFO: Deleting variant - variantID=%d", id)
	// Delete variant values first
	if err := s.variantRepo.DeleteVariantValuesByVariantID(id); err != nil {
		log.Printf("[productVariantService.DeleteVariant] ERROR: Failed to delete variant values - variantID=%d: %v", id, err)
		return fmt.Errorf("error deleting variant values: %w", err)
	}

	if err := s.variantRepo.DeleteVariant(id); err != nil {
		log.Printf("[productVariantService.DeleteVariant] ERROR: Failed to delete variant - variantID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantService.DeleteVariant] INFO: Variant deleted successfully - variantID=%d", id)
	return nil
}

// Variant Combination methods

func (s *productVariantService) CreateVariantCombination(productID uint, req *models.CreateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error) {
	log.Printf("[productVariantService.CreateVariantCombination] INFO: Creating combination - productID=%d, SKU=%s", productID, req.SKU)
	// Verify product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		log.Printf("[productVariantService.CreateVariantCombination] ERROR: Product not found - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	variantCombination, err := s.variantRepo.FindVariantCombinationBySKU(req.SKU)
	if err == nil && variantCombination.ProductID == productID {
		log.Printf("[productVariantService.CreateVariantCombination] INFO: Combination already exists - SKU=%s, combinationID=%d", req.SKU, variantCombination.ID)
		return s.toVariantCombinationResponse(variantCombination, product), nil
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
		log.Printf("[productVariantService.CreateVariantCombination] ERROR: Failed to create combination - SKU=%s: %v", req.SKU, err)
		return nil, fmt.Errorf("error creating variant combination: %w", err)
	}
	log.Printf("[productVariantService.CreateVariantCombination] INFO: Combination created successfully - combinationID=%d, SKU=%s", combination.ID, req.SKU)

	return s.toVariantCombinationResponse(combination, product), nil
}

func (s *productVariantService) GetVariantCombination(id uint) (*models.ProductVariantCombinationResponse, error) {
	log.Printf("[productVariantService.GetVariantCombination] INFO: Retrieving combination - combinationID=%d", id)
	combination, err := s.variantRepo.FindVariantCombinationByID(id)
	if err != nil {
		log.Printf("[productVariantService.GetVariantCombination] ERROR: Failed to find combination - combinationID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding variant combination: %w", err)
	}

	product, err := s.productRepo.FindByID(combination.ProductID)
	if err != nil {
		log.Printf("[productVariantService.GetVariantCombination] ERROR: Failed to find product for combination - productID=%d: %v", combination.ProductID, err)
		return nil, fmt.Errorf("error finding product: %w", err)
	}
	log.Printf("[productVariantService.GetVariantCombination] INFO: Combination found - combinationID=%d, SKU=%s", id, combination.SKU)

	return s.toVariantCombinationResponse(combination, product), nil
}

func (s *productVariantService) GetProductVariantCombinations(productID uint, limit, offset int) ([]models.ProductVariantCombinationResponse, error) {
	log.Printf("[productVariantService.GetProductVariantCombinations] INFO: Retrieving combinations for product - productID=%d, limit=%d, offset=%d", productID, limit, offset)
	// Verify product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		log.Printf("[productVariantService.GetProductVariantCombinations] ERROR: Product not found - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	combinations, err := s.variantRepo.FindVariantCombinationsByProductID(productID, limit, offset)
	if err != nil {
		log.Printf("[productVariantService.GetProductVariantCombinations] ERROR: Failed to find combinations - productID=%d: %v", productID, err)
		return nil, fmt.Errorf("error finding variant combinations: %w", err)
	}
	log.Printf("[productVariantService.GetProductVariantCombinations] INFO: Combinations retrieved successfully - productID=%d, count=%d", productID, len(combinations))

	var responses []models.ProductVariantCombinationResponse
	for _, combination := range combinations {
		responses = append(responses, *s.toVariantCombinationResponse(&combination, product))
	}

	return responses, nil
}

func (s *productVariantService) GetVariantCombinationBySKU(sku string) (*models.ProductVariantCombinationResponse, error) {
	log.Printf("[productVariantService.GetVariantCombinationBySKU] INFO: Retrieving combination - SKU=%s", sku)
	combination, err := s.variantRepo.FindVariantCombinationBySKU(sku)
	if err != nil {
		log.Printf("[productVariantService.GetVariantCombinationBySKU] ERROR: Failed to find combination - SKU=%s: %v", sku, err)
		return nil, fmt.Errorf("error finding variant combination: %w", err)
	}
	log.Printf("[productVariantService.GetVariantCombinationBySKU] INFO: Combination found - combinationID=%d, SKU=%s", combination.ID, sku)

	return s.toVariantCombinationResponse(combination, combination.Product), nil
}

func (s *productVariantService) UpdateVariantCombination(id uint, req *models.UpdateProductVariantCombinationRequest) (*models.ProductVariantCombinationResponse, error) {
	log.Printf("[productVariantService.UpdateVariantCombination] INFO: Updating combination - combinationID=%d", id)
	combination, err := s.variantRepo.FindVariantCombinationByID(id)
	if err != nil {
		log.Printf("[productVariantService.UpdateVariantCombination] ERROR: Failed to find combination - combinationID=%d: %v", id, err)
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
		log.Printf("[productVariantService.UpdateVariantCombination] ERROR: Failed to update combination - combinationID=%d: %v", id, err)
		return nil, fmt.Errorf("error updating variant combination: %w", err)
	}
	log.Printf("[productVariantService.UpdateVariantCombination] INFO: Combination updated successfully - combinationID=%d", id)

	return s.GetVariantCombination(id)
}

func (s *productVariantService) DeleteVariantCombination(id uint) error {
	log.Printf("[productVariantService.DeleteVariantCombination] INFO: Deleting combination - combinationID=%d", id)
	if err := s.variantRepo.DeleteVariantCombination(id); err != nil {
		log.Printf("[productVariantService.DeleteVariantCombination] ERROR: Failed to delete combination - combinationID=%d: %v", id, err)
		return err
	}
	log.Printf("[productVariantService.DeleteVariantCombination] INFO: Combination deleted successfully - combinationID=%d", id)
	return nil
}

// Helper functions

func (s *productVariantService) toVariantResponse(variant *models.ProductVariant, values []models.ProductVariantValue) *models.ProductVariantResponse {
	valueResponses := make([]models.ProductVariantValueResponse, len(values))
	for i, v := range values {
		valueResponses[i] = models.ProductVariantValueResponse{
			ID:    v.ID,
			Value: v.Value,
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
