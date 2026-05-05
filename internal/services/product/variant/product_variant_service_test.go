package variant

import (
	"encoding/json"
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductVariantRepository is a mock of ProductVariantRepository
type MockProductVariantRepository struct {
	mock.Mock
}

func (m *MockProductVariantRepository) CreateVariant(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) FindVariantByID(id uint) (*models.ProductVariant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantsByProductID(productID uint) ([]models.ProductVariant, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) UpdateVariant(variant *models.ProductVariant) error {
	args := m.Called(variant)
	return args.Error(0)
}

func (m *MockProductVariantRepository) DeleteVariant(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductVariantRepository) CreateVariantValue(value *models.ProductVariantValue) (*models.ProductVariantValue, error) {
	args := m.Called(value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantValue), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantsByProductIDAndValue(productID uint, value string) ([]models.ProductVariant, error) {
	args := m.Called(productID, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariant), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantValueByID(id uint) (*models.ProductVariantValue, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantValue), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantValuesByVariantID(variantID uint) ([]models.ProductVariantValue, error) {
	args := m.Called(variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariantValue), args.Error(1)
}

func (m *MockProductVariantRepository) UpdateVariantValue(value *models.ProductVariantValue) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *MockProductVariantRepository) DeleteVariantValue(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductVariantRepository) DeleteVariantValuesByVariantID(variantID uint) error {
	args := m.Called(variantID)
	return args.Error(0)
}

func (m *MockProductVariantRepository) CreateVariantCombination(combination *models.ProductVariantCombination) error {
	args := m.Called(combination)
	return args.Error(0)
}

func (m *MockProductVariantRepository) FindVariantCombinationByID(id uint) (*models.ProductVariantCombination, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantCombination), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantCombinationsByProductID(productID uint, limit, offset int) ([]models.ProductVariantCombination, error) {
	args := m.Called(productID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ProductVariantCombination), args.Error(1)
}

func (m *MockProductVariantRepository) FindVariantCombinationBySKU(sku string) (*models.ProductVariantCombination, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductVariantCombination), args.Error(1)
}

func (m *MockProductVariantRepository) UpdateVariantCombination(combination *models.ProductVariantCombination) error {
	args := m.Called(combination)
	return args.Error(0)
}

func (m *MockProductVariantRepository) DeleteVariantCombination(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductVariantRepository) UpdateVariantCombinationStock(id uint, quantity int) error {
	args := m.Called(id, quantity)
	return args.Error(0)
}

// MockProductRepository is a mock of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) FindByID(id uint) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) FindBySKU(sku string) (*models.Product, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) FindAll(limit, offset int) ([]models.Product, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) FindByCategory(category string, limit, offset int) ([]models.Product, error) {
	args := m.Called(category, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) FindActive(limit, offset int) ([]models.Product, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateStock(id uint, quantity int) error {
	args := m.Called(id, quantity)
	return args.Error(0)
}

// Tests

func TestCreateVariant(t *testing.T) {
	mockVariantRepo := new(MockProductVariantRepository)
	mockProductRepo := new(MockProductRepository)

	service := NewProductVariantService(mockVariantRepo, mockProductRepo)

	product := &models.Product{ID: 1, Name: "Test Product"}
	req := &models.CreateProductVariantRequest{
		Name:   "Color",
		Type:   "color",
		Values: []string{"Red", "Blue", "Green"},
	}

	mockProductRepo.On("FindByID", uint(1)).Return(product, nil)
	mockVariantRepo.On("FindVariantsByProductIDAndValue", uint(1), "Color").Return([]models.ProductVariant{}, nil)
	mockVariantRepo.On("CreateVariant", mock.MatchedBy(func(v *models.ProductVariant) bool {
		return v.ProductID == 1 && v.Name == "Color" && v.Type == "color"
	})).Run(func(args mock.Arguments) {
		variant := args.Get(0).(*models.ProductVariant)
		variant.ID = 1
	}).Return(nil)

	mockVariantRepo.On("CreateVariantValue", mock.MatchedBy(func(vv *models.ProductVariantValue) bool {
		return vv.VariantID == 1 && (vv.Value == "Red" || vv.Value == "Blue" || vv.Value == "Green")
	})).Return(&models.ProductVariantValue{}, nil)

	resp, err := service.CreateVariant(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Color", resp.Name)
	assert.Equal(t, "color", resp.Type)
	assert.Len(t, resp.Values, 3)

	mockProductRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
}

func TestCreateVariant_ExistingVariant(t *testing.T) {
	mockVariantRepo := new(MockProductVariantRepository)
	mockProductRepo := new(MockProductRepository)

	service := NewProductVariantService(mockVariantRepo, mockProductRepo)

	productID := uint(1)
	variantID := uint(10)
	product := &models.Product{ID: productID, Name: "Test Product"}

	existingVariant := models.ProductVariant{
		ID:        variantID,
		ProductID: productID,
		Name:      "Color",
		Type:      "color",
	}

	existingValues := []models.ProductVariantValue{
		{ID: 100, VariantID: variantID, Value: "Red"},
		{ID: 101, VariantID: variantID, Value: "Blue"},
	}

	req := &models.CreateProductVariantRequest{
		Name:   "Color",
		Type:   "color",
		Values: []string{"Red", "Green"}, // Red exists, Green is new
	}

	mockProductRepo.On("FindByID", productID).Return(product, nil)
	mockVariantRepo.On("FindVariantsByProductIDAndValue", productID, "Color").Return([]models.ProductVariant{existingVariant}, nil)
	mockVariantRepo.On("FindVariantValuesByVariantID", variantID).Return(existingValues, nil)

	// Expect only "Green" to be created
	mockVariantRepo.On("CreateVariantValue", mock.MatchedBy(func(vv *models.ProductVariantValue) bool {
		return vv.VariantID == variantID && vv.Value == "Green"
	})).Return(&models.ProductVariantValue{ID: 102, VariantID: variantID, Value: "Green"}, nil)

	resp, err := service.CreateVariant(productID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Color", resp.Name)
	// Response should contain Red, Blue (existing) and Green (new)
	var respValues []string
	for _, v := range resp.Values {
		respValues = append(respValues, v.Value)
	}
	assert.ElementsMatch(t, []string{"Red", "Blue", "Green"}, respValues)

	mockProductRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
}

func TestCreateVariantCombination(t *testing.T) {
	mockVariantRepo := new(MockProductVariantRepository)
	mockProductRepo := new(MockProductRepository)

	service := NewProductVariantService(mockVariantRepo, mockProductRepo)

	product := &models.Product{
		ID:          1,
		PriceRetail: 100.00,
		Name:        "Test Product",
	}
	req := &models.CreateProductVariantCombinationRequest{
		SKU: "TEST-RED-SMALL",
		VariantCombination: map[string]string{
			"Color": "Red",
			"Size":  "Small",
		},
		Stock:           50,
		PriceAdjustment: 10.00,
		ImageURL:        "https://example.com/image.jpg",
	}

	mockProductRepo.On("FindByID", uint(1)).Return(product, nil)
	mockVariantRepo.On("CreateVariantCombination", mock.MatchedBy(func(vc *models.ProductVariantCombination) bool {
		return vc.ProductID == 1 && vc.SKU == "TEST-RED-SMALL" && vc.Stock == 50
	})).Run(func(args mock.Arguments) {
		combination := args.Get(0).(*models.ProductVariantCombination)
		combination.ID = 1
	}).Return(nil)

	resp, err := service.CreateVariantCombination(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "TEST-RED-SMALL", resp.SKU)
	assert.Equal(t, 50, resp.Stock)
	assert.Equal(t, 10.00, resp.PriceAdjustment)
	assert.Equal(t, 110.00, resp.FinalPrice) // Base price + adjustment

	mockProductRepo.AssertExpectations(t)
	mockVariantRepo.AssertExpectations(t)
}

func TestUpdateVariantCombination(t *testing.T) {
	mockVariantRepo := new(MockProductVariantRepository)
	mockProductRepo := new(MockProductRepository)

	service := NewProductVariantService(mockVariantRepo, mockProductRepo)

	combinationData := map[string]string{
		"Color": "Red",
		"Size":  "Small",
	}
	combinationJSON, _ := json.Marshal(combinationData)

	combination := &models.ProductVariantCombination{
		ID:                 1,
		ProductID:          1,
		SKU:                "TEST-RED-SMALL",
		VariantCombination: string(combinationJSON),
		Stock:              50,
		PriceAdjustment:    10.00,
	}

	product := &models.Product{
		ID:          1,
		PriceRetail: 100.00,
		Name:        "Test Product",
	}

	newStock := 75
	newPriceAdjustment := 15.00

	req := &models.UpdateProductVariantCombinationRequest{
		Stock:           &newStock,
		PriceAdjustment: &newPriceAdjustment,
	}

	mockVariantRepo.On("FindVariantCombinationByID", uint(1)).Return(combination, nil)
	mockVariantRepo.On("UpdateVariantCombination", mock.MatchedBy(func(vc *models.ProductVariantCombination) bool {
		return vc.ID == 1 && vc.Stock == 75 && vc.PriceAdjustment == 15.00
	})).Return(nil)
	mockProductRepo.On("FindByID", uint(1)).Return(product, nil)
	mockVariantRepo.On("FindVariantCombinationByID", uint(1)).Return(combination, nil)

	resp, err := service.UpdateVariantCombination(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	mockVariantRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}
