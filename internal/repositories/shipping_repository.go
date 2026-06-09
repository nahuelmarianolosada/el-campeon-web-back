package repositories

import (
	"log"
	"strings"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ShippingRepository interface {
	// Zonas
	CreateZone(z *models.DeliveryZone) error
	UpdateZone(z *models.DeliveryZone) error
	DeleteZone(id uint) error
	FindZoneByID(id uint) (*models.DeliveryZone, error)
	ListZones(onlyActive bool) ([]models.DeliveryZone, error)

	// Tarifas
	CreateRate(r *models.DeliveryRate) error
	UpdateRate(r *models.DeliveryRate) error
	DeleteRate(id uint) error
	FindRateByID(id uint) (*models.DeliveryRate, error)
	FindActiveRatesForZone(zoneID uint) ([]models.DeliveryRate, error)
	FindRateByZoneAndBranch(zoneID, branchID uint) (*models.DeliveryRate, error)
	ListRates(zoneID, branchID *uint) ([]models.DeliveryRate, error)

	// Códigos postales
	UpsertPostalCode(pc *models.PostalCodeZone) error
	BulkUpsertPostalCodes(entries []models.BulkPostalCodeEntry) error
	DeletePostalCode(postalCode string) error
	ListPostalCodes(zoneID *uint) ([]models.PostalCodeZone, error)
	FindZoneByPostalCode(postalCode string) (*models.DeliveryZone, error)
}

type shippingRepository struct {
	db *gorm.DB
}

func NewShippingRepository(db *gorm.DB) ShippingRepository {
	return &shippingRepository{db: db}
}

// ----- Zonas -----

func (r *shippingRepository) CreateZone(z *models.DeliveryZone) error {
	log.Printf("[shippingRepository.CreateZone] INFO: name=%s kind=%s", z.Name, z.Kind)
	return r.db.Create(z).Error
}

func (r *shippingRepository) UpdateZone(z *models.DeliveryZone) error {
	return r.db.Save(z).Error
}

func (r *shippingRepository) DeleteZone(id uint) error {
	return r.db.Delete(&models.DeliveryZone{}, id).Error
}

func (r *shippingRepository) FindZoneByID(id uint) (*models.DeliveryZone, error) {
	var z models.DeliveryZone
	if err := r.db.First(&z, id).Error; err != nil {
		return nil, err
	}
	return &z, nil
}

func (r *shippingRepository) ListZones(onlyActive bool) ([]models.DeliveryZone, error) {
	var zones []models.DeliveryZone
	q := r.db.Model(&models.DeliveryZone{})
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("name ASC").Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

// ----- Tarifas -----

func (r *shippingRepository) CreateRate(rate *models.DeliveryRate) error {
	log.Printf("[shippingRepository.CreateRate] INFO: zoneID=%d branchID=%d cost=%.2f", rate.ZoneID, rate.OriginBranchID, rate.Cost)
	return r.db.Create(rate).Error
}

func (r *shippingRepository) UpdateRate(rate *models.DeliveryRate) error {
	return r.db.Save(rate).Error
}

func (r *shippingRepository) DeleteRate(id uint) error {
	return r.db.Delete(&models.DeliveryRate{}, id).Error
}

func (r *shippingRepository) FindRateByID(id uint) (*models.DeliveryRate, error) {
	var rate models.DeliveryRate
	if err := r.db.First(&rate, id).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *shippingRepository) FindActiveRatesForZone(zoneID uint) ([]models.DeliveryRate, error) {
	var rates []models.DeliveryRate
	if err := r.db.
		Where("zone_id = ? AND is_active = ?", zoneID, true).
		Order("cost ASC").
		Find(&rates).Error; err != nil {
		return nil, err
	}
	return rates, nil
}

func (r *shippingRepository) FindRateByZoneAndBranch(zoneID, branchID uint) (*models.DeliveryRate, error) {
	var rate models.DeliveryRate
	if err := r.db.
		Where("zone_id = ? AND origin_branch_id = ? AND is_active = ?", zoneID, branchID, true).
		First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *shippingRepository) ListRates(zoneID, branchID *uint) ([]models.DeliveryRate, error) {
	var rates []models.DeliveryRate
	q := r.db.Model(&models.DeliveryRate{})
	if zoneID != nil {
		q = q.Where("zone_id = ?", *zoneID)
	}
	if branchID != nil {
		q = q.Where("origin_branch_id = ?", *branchID)
	}
	if err := q.Order("zone_id ASC, origin_branch_id ASC").Find(&rates).Error; err != nil {
		return nil, err
	}
	return rates, nil
}

// ----- Códigos postales -----

func normalizePostalCode(pc string) string {
	return strings.ToUpper(strings.TrimSpace(pc))
}

func (r *shippingRepository) UpsertPostalCode(pc *models.PostalCodeZone) error {
	pc.PostalCode = normalizePostalCode(pc.PostalCode)
	sql := `INSERT INTO postal_code_zones (postal_code, zone_id) VALUES (?, ?)
	        ON DUPLICATE KEY UPDATE zone_id = VALUES(zone_id)`
	return r.db.Exec(sql, pc.PostalCode, pc.ZoneID).Error
}

func (r *shippingRepository) BulkUpsertPostalCodes(entries []models.BulkPostalCodeEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx := r.db.Begin()
	for _, e := range entries {
		code := normalizePostalCode(e.PostalCode)
		if err := tx.Exec(
			`INSERT INTO postal_code_zones (postal_code, zone_id) VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE zone_id = VALUES(zone_id)`,
			code, e.ZoneID,
		).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *shippingRepository) DeletePostalCode(postalCode string) error {
	return r.db.Where("postal_code = ?", normalizePostalCode(postalCode)).
		Delete(&models.PostalCodeZone{}).Error
}

func (r *shippingRepository) ListPostalCodes(zoneID *uint) ([]models.PostalCodeZone, error) {
	var pcs []models.PostalCodeZone
	q := r.db.Model(&models.PostalCodeZone{})
	if zoneID != nil {
		q = q.Where("zone_id = ?", *zoneID)
	}
	if err := q.Order("postal_code ASC").Find(&pcs).Error; err != nil {
		return nil, err
	}
	return pcs, nil
}

func (r *shippingRepository) FindZoneByPostalCode(postalCode string) (*models.DeliveryZone, error) {
	var z models.DeliveryZone
	err := r.db.
		Joins("JOIN postal_code_zones pcz ON pcz.zone_id = delivery_zones.id").
		Where("pcz.postal_code = ? AND delivery_zones.is_active = ?", normalizePostalCode(postalCode), true).
		First(&z).Error
	if err != nil {
		return nil, err
	}
	return &z, nil
}
