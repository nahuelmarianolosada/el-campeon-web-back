package shipping

import (
	"fmt"
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type ShippingService interface {
	// Branches
	CreateBranch(req *models.CreateBranchRequest) (*models.BranchResponse, error)
	UpdateBranch(id uint, req *models.UpdateBranchRequest) (*models.BranchResponse, error)
	DeleteBranch(id uint) error
	GetBranch(id uint) (*models.BranchResponse, error)
	ListBranches(onlyActive, onlyPickup bool) ([]models.BranchResponse, error)

	// Zonas
	CreateZone(req *models.CreateZoneRequest) (*models.DeliveryZone, error)
	UpdateZone(id uint, req *models.UpdateZoneRequest) (*models.DeliveryZone, error)
	DeleteZone(id uint) error
	ListZones(onlyActive bool) ([]models.DeliveryZone, error)

	// Tarifas
	CreateRate(req *models.CreateRateRequest) (*models.DeliveryRate, error)
	UpdateRate(id uint, req *models.UpdateRateRequest) (*models.DeliveryRate, error)
	DeleteRate(id uint) error
	ListRates(zoneID, branchID *uint) ([]models.DeliveryRate, error)

	// Códigos postales
	UpsertPostalCode(req *models.UpsertPostalCodeRequest) error
	BulkUpsertPostalCodes(req *models.BulkPostalCodeRequest) error
	DeletePostalCode(postalCode string) error
	ListPostalCodes(zoneID *uint) ([]models.PostalCodeZone, error)

	// Stock por sucursal
	GetProductStock(productID uint) ([]models.ProductBranchStockResponse, error)
	SetProductStock(productID uint, req *models.UpdateBranchStockRequest) error

	// Cotización (público, usado en checkout)
	Quote(req *models.ShippingQuoteRequest) (*models.ShippingQuoteResponse, error)
}

type shippingService struct {
	branchRepo   repositories.BranchRepository
	shippingRepo repositories.ShippingRepository
	stockRepo    repositories.ProductBranchStockRepository
}

func NewShippingService(
	branchRepo repositories.BranchRepository,
	shippingRepo repositories.ShippingRepository,
	stockRepo repositories.ProductBranchStockRepository,
) ShippingService {
	return &shippingService{
		branchRepo:   branchRepo,
		shippingRepo: shippingRepo,
		stockRepo:    stockRepo,
	}
}

// ===== Branches =====

func (s *shippingService) CreateBranch(req *models.CreateBranchRequest) (*models.BranchResponse, error) {
	b := &models.Branch{
		Code:          req.Code,
		Name:          req.Name,
		Address:       req.Address,
		Lat:           req.Lat,
		Lng:           req.Lng,
		IsPickupPoint: derefBool(req.IsPickupPoint, true),
		IsActive:      derefBool(req.IsActive, true),
	}
	if err := s.branchRepo.Create(b); err != nil {
		return nil, fmt.Errorf("error creating branch: %w", err)
	}
	resp := b.ToResponse()
	return &resp, nil
}

func (s *shippingService) UpdateBranch(id uint, req *models.UpdateBranchRequest) (*models.BranchResponse, error) {
	b, err := s.branchRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("branch not found: %w", err)
	}
	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.Address != nil {
		b.Address = *req.Address
	}
	if req.Lat != nil {
		b.Lat = req.Lat
	}
	if req.Lng != nil {
		b.Lng = req.Lng
	}
	if req.IsPickupPoint != nil {
		b.IsPickupPoint = *req.IsPickupPoint
	}
	if req.IsActive != nil {
		b.IsActive = *req.IsActive
	}
	if err := s.branchRepo.Update(b); err != nil {
		return nil, err
	}
	resp := b.ToResponse()
	return &resp, nil
}

func (s *shippingService) DeleteBranch(id uint) error {
	return s.branchRepo.Delete(id)
}

func (s *shippingService) GetBranch(id uint) (*models.BranchResponse, error) {
	b, err := s.branchRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := b.ToResponse()
	return &resp, nil
}

func (s *shippingService) ListBranches(onlyActive, onlyPickup bool) ([]models.BranchResponse, error) {
	bs, err := s.branchRepo.FindAll(onlyActive, onlyPickup)
	if err != nil {
		return nil, err
	}
	out := make([]models.BranchResponse, 0, len(bs))
	for i := range bs {
		out = append(out, bs[i].ToResponse())
	}
	return out, nil
}

// ===== Zonas =====

func (s *shippingService) CreateZone(req *models.CreateZoneRequest) (*models.DeliveryZone, error) {
	z := &models.DeliveryZone{
		Name:         req.Name,
		Kind:         req.Kind,
		ParentZoneID: req.ParentZoneID,
		IsActive:     derefBool(req.IsActive, true),
	}
	if err := s.shippingRepo.CreateZone(z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *shippingService) UpdateZone(id uint, req *models.UpdateZoneRequest) (*models.DeliveryZone, error) {
	z, err := s.shippingRepo.FindZoneByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		z.Name = *req.Name
	}
	if req.Kind != nil {
		z.Kind = *req.Kind
	}
	if req.ParentZoneID != nil {
		z.ParentZoneID = req.ParentZoneID
	}
	if req.IsActive != nil {
		z.IsActive = *req.IsActive
	}
	if err := s.shippingRepo.UpdateZone(z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *shippingService) DeleteZone(id uint) error {
	return s.shippingRepo.DeleteZone(id)
}

func (s *shippingService) ListZones(onlyActive bool) ([]models.DeliveryZone, error) {
	return s.shippingRepo.ListZones(onlyActive)
}

// ===== Tarifas =====

func (s *shippingService) CreateRate(req *models.CreateRateRequest) (*models.DeliveryRate, error) {
	r := &models.DeliveryRate{
		ZoneID:                req.ZoneID,
		OriginBranchID:        req.OriginBranchID,
		Cost:                  req.Cost,
		EtaMinDays:            req.EtaMinDays,
		EtaMaxDays:            req.EtaMaxDays,
		FreeShippingThreshold: req.FreeShippingThreshold,
		IsActive:              derefBool(req.IsActive, true),
	}
	if r.EtaMaxDays < r.EtaMinDays {
		return nil, fmt.Errorf("eta_max_days must be >= eta_min_days")
	}
	if err := s.shippingRepo.CreateRate(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *shippingService) UpdateRate(id uint, req *models.UpdateRateRequest) (*models.DeliveryRate, error) {
	r, err := s.shippingRepo.FindRateByID(id)
	if err != nil {
		return nil, err
	}
	if req.Cost != nil {
		r.Cost = *req.Cost
	}
	if req.EtaMinDays != nil {
		r.EtaMinDays = *req.EtaMinDays
	}
	if req.EtaMaxDays != nil {
		r.EtaMaxDays = *req.EtaMaxDays
	}
	if req.FreeShippingThreshold != nil {
		r.FreeShippingThreshold = req.FreeShippingThreshold
	}
	if req.IsActive != nil {
		r.IsActive = *req.IsActive
	}
	if r.EtaMaxDays < r.EtaMinDays {
		return nil, fmt.Errorf("eta_max_days must be >= eta_min_days")
	}
	if err := s.shippingRepo.UpdateRate(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *shippingService) DeleteRate(id uint) error {
	return s.shippingRepo.DeleteRate(id)
}

func (s *shippingService) ListRates(zoneID, branchID *uint) ([]models.DeliveryRate, error) {
	return s.shippingRepo.ListRates(zoneID, branchID)
}

// ===== Códigos postales =====

func (s *shippingService) UpsertPostalCode(req *models.UpsertPostalCodeRequest) error {
	return s.shippingRepo.UpsertPostalCode(&models.PostalCodeZone{
		PostalCode: req.PostalCode,
		ZoneID:     req.ZoneID,
	})
}

func (s *shippingService) BulkUpsertPostalCodes(req *models.BulkPostalCodeRequest) error {
	return s.shippingRepo.BulkUpsertPostalCodes(req.Entries)
}

func (s *shippingService) DeletePostalCode(postalCode string) error {
	return s.shippingRepo.DeletePostalCode(postalCode)
}

func (s *shippingService) ListPostalCodes(zoneID *uint) ([]models.PostalCodeZone, error) {
	return s.shippingRepo.ListPostalCodes(zoneID)
}

// ===== Stock por sucursal =====

func (s *shippingService) GetProductStock(productID uint) ([]models.ProductBranchStockResponse, error) {
	rows, err := s.stockRepo.FindByProduct(productID)
	if err != nil {
		return nil, err
	}
	branches, err := s.branchRepo.FindAll(false, false)
	if err != nil {
		return nil, err
	}
	branchByID := make(map[uint]models.Branch, len(branches))
	for _, b := range branches {
		branchByID[b.ID] = b
	}
	out := make([]models.ProductBranchStockResponse, 0, len(rows))
	for _, r := range rows {
		b := branchByID[r.BranchID]
		out = append(out, models.ProductBranchStockResponse{
			ProductID:  r.ProductID,
			BranchID:   r.BranchID,
			BranchCode: b.Code,
			BranchName: b.Name,
			Stock:      r.Stock,
			Reserved:   r.Reserved,
			Available:  r.Stock - r.Reserved,
		})
	}
	return out, nil
}

func (s *shippingService) SetProductStock(productID uint, req *models.UpdateBranchStockRequest) error {
	return s.stockRepo.Upsert(&models.ProductBranchStock{
		ProductID: productID,
		BranchID:  req.BranchID,
		Stock:     req.Stock,
	})
}

// ===== Cotización =====

// Quote resuelve la zona desde el CP, elige sucursal origen con stock completo,
// aplica el umbral de envío gratis y devuelve el costo + ETA.
func (s *shippingService) Quote(req *models.ShippingQuoteRequest) (*models.ShippingQuoteResponse, error) {
	log.Printf("[shippingService.Quote] INFO: postal_code=%s subtotal=%.2f items=%d", req.PostalCode, req.Subtotal, len(req.Items))

	// 1. Resolver zona por CP.
	zone, err := s.shippingRepo.FindZoneByPostalCode(req.PostalCode)
	if err != nil {
		log.Printf("[shippingService.Quote] WARN: postal_code=%s no cobertura: %v", req.PostalCode, err)
		return nil, ErrPostalCodeNotCovered
	}

	// 2. Buscar sucursales que tengan stock completo para todos los items.
	branchIDs, err := s.stockRepo.BranchesWithFullStock(req.Items)
	if err != nil {
		return nil, fmt.Errorf("error checking branch stock: %w", err)
	}
	if len(branchIDs) == 0 {
		// Devolvemos cuáles items faltan en ambas sucursales (informativo).
		missing := s.aggregateMissingAcrossBranches(req.Items)
		// Igual buscamos una tarifa razonable para mostrar el costo proyectado.
		return s.buildOutOfStockResponse(zone, req, missing)
	}

	// 3. Filtrar a sucursales activas con tarifa activa para la zona.
	rates, err := s.shippingRepo.FindActiveRatesForZone(zone.ID)
	if err != nil {
		return nil, fmt.Errorf("error loading rates: %w", err)
	}
	ratesByBranch := make(map[uint]models.DeliveryRate, len(rates))
	for _, r := range rates {
		ratesByBranch[r.OriginBranchID] = r
	}

	type candidate struct {
		branch models.Branch
		rate   models.DeliveryRate
	}
	var candidates []candidate
	for _, bid := range branchIDs {
		rate, ok := ratesByBranch[bid]
		if !ok {
			continue
		}
		b, err := s.branchRepo.FindByID(bid)
		if err != nil || !b.IsActive {
			continue
		}
		candidates = append(candidates, candidate{branch: *b, rate: rate})
	}
	if len(candidates) == 0 {
		return nil, ErrNoRateForZone
	}

	// 4. Elegir la más barata (regla simple — el admin puede sumar ETA/distancia luego).
	chosen := candidates[0]
	for _, c := range candidates[1:] {
		if c.rate.Cost < chosen.rate.Cost {
			chosen = c
		}
	}

	// 5. Aplicar umbral de envío gratis.
	cost := chosen.rate.Cost
	freeApplied := false
	var amountForFree *float64
	if chosen.rate.FreeShippingThreshold != nil {
		threshold := *chosen.rate.FreeShippingThreshold
		if req.Subtotal >= threshold {
			cost = 0
			freeApplied = true
		} else {
			d := threshold - req.Subtotal
			amountForFree = &d
		}
	}

	return &models.ShippingQuoteResponse{
		Zone: models.ShippingQuoteZone{
			ID:   zone.ID,
			Name: zone.Name,
			Kind: zone.Kind,
		},
		OriginBranchID:       chosen.branch.ID,
		OriginBranchName:     chosen.branch.Name,
		Cost:                 cost,
		EtaMinDays:           chosen.rate.EtaMinDays,
		EtaMaxDays:           chosen.rate.EtaMaxDays,
		FreeShippingApplied:  freeApplied,
		AmountForFreeShip:    amountForFree,
		InStock:              true,
		OutOfStockProductIDs: nil,
	}, nil
}

func (s *shippingService) buildOutOfStockResponse(
	zone *models.DeliveryZone,
	req *models.ShippingQuoteRequest,
	missing []uint,
) (*models.ShippingQuoteResponse, error) {
	// Buscamos una tarifa de la zona para devolver el costo proyectado, aunque
	// in_stock=false bloquee la compra en el front.
	rates, err := s.shippingRepo.FindActiveRatesForZone(zone.ID)
	if err != nil || len(rates) == 0 {
		return nil, ErrNoBranchHasStock
	}
	r := rates[0]
	return &models.ShippingQuoteResponse{
		Zone: models.ShippingQuoteZone{
			ID:   zone.ID,
			Name: zone.Name,
			Kind: zone.Kind,
		},
		OriginBranchID:       r.OriginBranchID,
		Cost:                 r.Cost,
		EtaMinDays:           r.EtaMinDays,
		EtaMaxDays:           r.EtaMaxDays,
		FreeShippingApplied:  false,
		InStock:              false,
		OutOfStockProductIDs: missing,
	}, nil
}

// aggregateMissingAcrossBranches devuelve los product_ids que ninguna sucursal puede cubrir.
func (s *shippingService) aggregateMissingAcrossBranches(items []models.QuoteItem) []uint {
	branches, _ := s.branchRepo.FindAll(true, false)
	if len(branches) == 0 {
		ids := make([]uint, 0, len(items))
		for _, it := range items {
			ids = append(ids, it.ProductID)
		}
		return ids
	}
	covered := make(map[uint]bool)
	for _, b := range branches {
		missing, _ := s.stockRepo.OutOfStockItemsForBranch(b.ID, items)
		missingSet := make(map[uint]bool, len(missing))
		for _, m := range missing {
			missingSet[m] = true
		}
		for _, it := range items {
			if !missingSet[it.ProductID] {
				covered[it.ProductID] = true
			}
		}
	}
	var missing []uint
	for _, it := range items {
		if !covered[it.ProductID] {
			missing = append(missing, it.ProductID)
		}
	}
	return missing
}

func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}
