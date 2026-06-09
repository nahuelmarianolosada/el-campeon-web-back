package shipping

import (
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ============ Mocks ============

type MockBranchRepository struct {
	mock.Mock
}

func (m *MockBranchRepository) Create(branch *models.Branch) error {
	args := m.Called(branch)
	branch.ID = 1
	return args.Error(0)
}
func (m *MockBranchRepository) FindByID(id uint) (*models.Branch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Branch), args.Error(1)
}
func (m *MockBranchRepository) FindByCode(code string) (*models.Branch, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Branch), args.Error(1)
}
func (m *MockBranchRepository) Update(branch *models.Branch) error {
	args := m.Called(branch)
	return args.Error(0)
}
func (m *MockBranchRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockBranchRepository) FindAll(onlyActive, onlyPickup bool) ([]models.Branch, error) {
	args := m.Called(onlyActive, onlyPickup)
	return args.Get(0).([]models.Branch), args.Error(1)
}

type MockShippingRepository struct {
	mock.Mock
}

func (m *MockShippingRepository) CreateZone(z *models.DeliveryZone) error {
	args := m.Called(z)
	z.ID = 1
	return args.Error(0)
}
func (m *MockShippingRepository) UpdateZone(z *models.DeliveryZone) error {
	args := m.Called(z)
	return args.Error(0)
}
func (m *MockShippingRepository) DeleteZone(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockShippingRepository) FindZoneByID(id uint) (*models.DeliveryZone, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeliveryZone), args.Error(1)
}
func (m *MockShippingRepository) ListZones(onlyActive bool) ([]models.DeliveryZone, error) {
	args := m.Called(onlyActive)
	return args.Get(0).([]models.DeliveryZone), args.Error(1)
}
func (m *MockShippingRepository) CreateRate(r *models.DeliveryRate) error {
	args := m.Called(r)
	r.ID = 1
	return args.Error(0)
}
func (m *MockShippingRepository) UpdateRate(r *models.DeliveryRate) error {
	args := m.Called(r)
	return args.Error(0)
}
func (m *MockShippingRepository) DeleteRate(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockShippingRepository) FindRateByID(id uint) (*models.DeliveryRate, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeliveryRate), args.Error(1)
}
func (m *MockShippingRepository) FindActiveRatesForZone(zoneID uint) ([]models.DeliveryRate, error) {
	args := m.Called(zoneID)
	return args.Get(0).([]models.DeliveryRate), args.Error(1)
}
func (m *MockShippingRepository) FindRateByZoneAndBranch(zoneID, branchID uint) (*models.DeliveryRate, error) {
	args := m.Called(zoneID, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeliveryRate), args.Error(1)
}
func (m *MockShippingRepository) ListRates(zoneID, branchID *uint) ([]models.DeliveryRate, error) {
	args := m.Called(zoneID, branchID)
	return args.Get(0).([]models.DeliveryRate), args.Error(1)
}
func (m *MockShippingRepository) UpsertPostalCode(pc *models.PostalCodeZone) error {
	args := m.Called(pc)
	return args.Error(0)
}
func (m *MockShippingRepository) BulkUpsertPostalCodes(entries []models.BulkPostalCodeEntry) error {
	args := m.Called(entries)
	return args.Error(0)
}
func (m *MockShippingRepository) DeletePostalCode(postalCode string) error {
	args := m.Called(postalCode)
	return args.Error(0)
}
func (m *MockShippingRepository) ListPostalCodes(zoneID *uint) ([]models.PostalCodeZone, error) {
	args := m.Called(zoneID)
	return args.Get(0).([]models.PostalCodeZone), args.Error(1)
}
func (m *MockShippingRepository) FindZoneByPostalCode(postalCode string) (*models.DeliveryZone, error) {
	args := m.Called(postalCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DeliveryZone), args.Error(1)
}

type MockProductBranchStockRepository struct {
	mock.Mock
}

func (m *MockProductBranchStockRepository) FindByProduct(productID uint) ([]models.ProductBranchStock, error) {
	args := m.Called(productID)
	return args.Get(0).([]models.ProductBranchStock), args.Error(1)
}
func (m *MockProductBranchStockRepository) GetByProductAndBranch(productID, branchID uint) (*models.ProductBranchStock, error) {
	args := m.Called(productID, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductBranchStock), args.Error(1)
}
func (m *MockProductBranchStockRepository) Upsert(stock *models.ProductBranchStock) error {
	args := m.Called(stock)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) IncrementReserved(tx *gorm.DB, productID, branchID uint, delta int) error {
	args := m.Called(tx, productID, branchID, delta)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) DecrementStock(tx *gorm.DB, productID, branchID uint, qty int) error {
	args := m.Called(tx, productID, branchID, qty)
	return args.Error(0)
}
func (m *MockProductBranchStockRepository) BranchesWithFullStock(items []models.QuoteItem) ([]uint, error) {
	args := m.Called(items)
	return args.Get(0).([]uint), args.Error(1)
}
func (m *MockProductBranchStockRepository) OutOfStockItemsForBranch(branchID uint, items []models.QuoteItem) ([]uint, error) {
	args := m.Called(branchID, items)
	return args.Get(0).([]uint), args.Error(1)
}
func (m *MockProductBranchStockRepository) DB() *gorm.DB {
	return nil
}

// ============ Tests ============

func TestShippingService_CreateBranch(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	service := NewShippingService(branchRepo, nil, nil)

	req := &models.CreateBranchRequest{
		Code:    "B1",
		Name:    "Branch 1",
		Address: "Address 1",
	}

	branchRepo.On("Create", mock.AnythingOfType("*models.Branch")).Return(nil)

	resp, err := service.CreateBranch(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "B1", resp.Code)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_Quote_Success(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	shippingRepo := new(MockShippingRepository)
	stockRepo := new(MockProductBranchStockRepository)
	service := NewShippingService(branchRepo, shippingRepo, stockRepo)

	req := &models.ShippingQuoteRequest{
		PostalCode: "1234",
		Subtotal:   1000.0,
		Items: []models.QuoteItem{
			{ProductID: 1, Quantity: 2},
		},
	}

	zone := &models.DeliveryZone{ID: 1, Name: "Zone 1", Kind: "provincial"}
	shippingRepo.On("FindZoneByPostalCode", "1234").Return(zone, nil)

	stockRepo.On("BranchesWithFullStock", req.Items).Return([]uint{1}, nil)

	rate := models.DeliveryRate{
		OriginBranchID:        1,
		Cost:                  100.0,
		EtaMinDays:            1,
		EtaMaxDays:            3,
		FreeShippingThreshold: float64Ptr(1500.0),
	}
	shippingRepo.On("FindActiveRatesForZone", uint(1)).Return([]models.DeliveryRate{rate}, nil)

	branch := &models.Branch{ID: 1, Name: "Branch 1", IsActive: true}
	branchRepo.On("FindByID", uint(1)).Return(branch, nil)

	resp, err := service.Quote(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 100.0, resp.Cost)
	assert.False(t, resp.FreeShippingApplied)
	assert.Equal(t, 500.0, *resp.AmountForFreeShip)
	assert.True(t, resp.InStock)

	shippingRepo.AssertExpectations(t)
	stockRepo.AssertExpectations(t)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_Quote_FreeShipping(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	shippingRepo := new(MockShippingRepository)
	stockRepo := new(MockProductBranchStockRepository)
	service := NewShippingService(branchRepo, shippingRepo, stockRepo)

	req := &models.ShippingQuoteRequest{
		PostalCode: "1234",
		Subtotal:   2000.0,
		Items: []models.QuoteItem{
			{ProductID: 1, Quantity: 2},
		},
	}

	zone := &models.DeliveryZone{ID: 1, Name: "Zone 1", Kind: "provincial"}
	shippingRepo.On("FindZoneByPostalCode", "1234").Return(zone, nil)

	stockRepo.On("BranchesWithFullStock", req.Items).Return([]uint{1}, nil)

	rate := models.DeliveryRate{
		OriginBranchID:        1,
		Cost:                  100.0,
		EtaMinDays:            1,
		EtaMaxDays:            3,
		FreeShippingThreshold: float64Ptr(1500.0),
	}
	shippingRepo.On("FindActiveRatesForZone", uint(1)).Return([]models.DeliveryRate{rate}, nil)

	branch := &models.Branch{ID: 1, Name: "Branch 1", IsActive: true}
	branchRepo.On("FindByID", uint(1)).Return(branch, nil)

	resp, err := service.Quote(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0.0, resp.Cost)
	assert.True(t, resp.FreeShippingApplied)
	assert.Nil(t, resp.AmountForFreeShip)

	shippingRepo.AssertExpectations(t)
	stockRepo.AssertExpectations(t)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_Quote_OutOfStock(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	shippingRepo := new(MockShippingRepository)
	stockRepo := new(MockProductBranchStockRepository)
	service := NewShippingService(branchRepo, shippingRepo, stockRepo)

	req := &models.ShippingQuoteRequest{
		PostalCode: "1234",
		Subtotal:   1000.0,
		Items: []models.QuoteItem{
			{ProductID: 1, Quantity: 2},
		},
	}

	zone := &models.DeliveryZone{ID: 1, Name: "Zone 1", Kind: "provincial"}
	shippingRepo.On("FindZoneByPostalCode", "1234").Return(zone, nil)

	stockRepo.On("BranchesWithFullStock", req.Items).Return([]uint{}, nil)

	// Mock aggregateMissingAcrossBranches
	branch1 := models.Branch{ID: 1, IsActive: true}
	branchRepo.On("FindAll", true, false).Return([]models.Branch{branch1}, nil)
	stockRepo.On("OutOfStockItemsForBranch", uint(1), req.Items).Return([]uint{1}, nil)

	// buildOutOfStockResponse
	rate := models.DeliveryRate{
		OriginBranchID: 1,
		Cost:           100.0,
		EtaMinDays:     1,
		EtaMaxDays:     3,
	}
	shippingRepo.On("FindActiveRatesForZone", uint(1)).Return([]models.DeliveryRate{rate}, nil)

	resp, err := service.Quote(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.InStock)
	assert.Equal(t, []uint{1}, resp.OutOfStockProductIDs)

	shippingRepo.AssertExpectations(t)
	stockRepo.AssertExpectations(t)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_UpdateBranch(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	service := NewShippingService(branchRepo, nil, nil)

	branch := &models.Branch{ID: 1, Name: "Old Name"}
	branchRepo.On("FindByID", uint(1)).Return(branch, nil)
	branchRepo.On("Update", mock.Anything).Return(nil)

	newName := "New Name"
	req := &models.UpdateBranchRequest{Name: &newName}
	resp, err := service.UpdateBranch(1, req)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", resp.Name)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_DeleteBranch(t *testing.T) {
	branchRepo := new(MockBranchRepository)
	service := NewShippingService(branchRepo, nil, nil)

	branchRepo.On("Delete", uint(1)).Return(nil)

	err := service.DeleteBranch(1)

	assert.NoError(t, err)
	branchRepo.AssertExpectations(t)
}

func TestShippingService_CreateZone(t *testing.T) {
	shippingRepo := new(MockShippingRepository)
	service := NewShippingService(nil, shippingRepo, nil)

	req := &models.CreateZoneRequest{
		Name: "Zone A",
		Kind: "provincial",
	}
	shippingRepo.On("CreateZone", mock.Anything).Return(nil)

	resp, err := service.CreateZone(req)

	assert.NoError(t, err)
	assert.Equal(t, "Zone A", resp.Name)
	shippingRepo.AssertExpectations(t)
}

func TestShippingService_CreateRate_InvalidEta(t *testing.T) {
	service := NewShippingService(nil, nil, nil)

	req := &models.CreateRateRequest{
		EtaMinDays: 5,
		EtaMaxDays: 2,
	}

	_, err := service.CreateRate(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "eta_max_days must be >= eta_min_days")
}

func float64Ptr(f float64) *float64 {
	return &f
}
