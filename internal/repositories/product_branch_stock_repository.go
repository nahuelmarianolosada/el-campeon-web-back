package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ProductBranchStockRepository interface {
	FindByProduct(productID uint) ([]models.ProductBranchStock, error)
	GetByProductAndBranch(productID, branchID uint) (*models.ProductBranchStock, error)
	Upsert(stock *models.ProductBranchStock) error
	IncrementReserved(tx *gorm.DB, productID, branchID uint, delta int) error
	DecrementStock(tx *gorm.DB, productID, branchID uint, qty int) error
	BranchesWithFullStock(items []models.QuoteItem) ([]uint, error)
	OutOfStockItemsForBranch(branchID uint, items []models.QuoteItem) ([]uint, error)
	DB() *gorm.DB
}

type productBranchStockRepository struct {
	db *gorm.DB
}

func NewProductBranchStockRepository(db *gorm.DB) ProductBranchStockRepository {
	return &productBranchStockRepository{db: db}
}

func (r *productBranchStockRepository) DB() *gorm.DB { return r.db }

func (r *productBranchStockRepository) FindByProduct(productID uint) ([]models.ProductBranchStock, error) {
	var rows []models.ProductBranchStock
	if err := r.db.Where("product_id = ?", productID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *productBranchStockRepository) GetByProductAndBranch(productID, branchID uint) (*models.ProductBranchStock, error) {
	var pbs models.ProductBranchStock
	if err := r.db.Where("product_id = ? AND branch_id = ?", productID, branchID).First(&pbs).Error; err != nil {
		return nil, err
	}
	return &pbs, nil
}

// Upsert inserta una fila o actualiza el stock si ya existe.
func (r *productBranchStockRepository) Upsert(stock *models.ProductBranchStock) error {
	log.Printf("[productBranchStockRepository.Upsert] INFO: productID=%d, branchID=%d, stock=%d", stock.ProductID, stock.BranchID, stock.Stock)
	sql := `
		INSERT INTO product_branch_stock (product_id, branch_id, stock, reserved)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE stock = VALUES(stock), reserved = VALUES(reserved)`
	return r.db.Exec(sql, stock.ProductID, stock.BranchID, stock.Stock, stock.Reserved).Error
}

func (r *productBranchStockRepository) IncrementReserved(tx *gorm.DB, productID, branchID uint, delta int) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.Model(&models.ProductBranchStock{}).
		Where("product_id = ? AND branch_id = ?", productID, branchID).
		UpdateColumn("reserved", gorm.Expr("reserved + ?", delta)).Error
}

func (r *productBranchStockRepository) DecrementStock(tx *gorm.DB, productID, branchID uint, qty int) error {
	db := tx
	if db == nil {
		db = r.db
	}
	res := db.Exec(
		"UPDATE product_branch_stock SET stock = stock - ? WHERE product_id = ? AND branch_id = ? AND stock >= ?",
		qty, productID, branchID, qty,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BranchesWithFullStock devuelve los IDs de sucursales (activas) que tienen
// stock suficiente para CADA item del request, en una sola consulta.
func (r *productBranchStockRepository) BranchesWithFullStock(items []models.QuoteItem) ([]uint, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Sumamos coberturas: una sucursal cubre un item si stock - reserved >= qty.
	// Una sucursal "cubre todo" si COUNT(items cubiertos) = len(items).
	type row struct {
		BranchID uint
		Covered  int
	}
	var rows []row

	// Construyo lista de productos y mapa de cantidades requeridas.
	productIDs := make([]uint, 0, len(items))
	qtyByProduct := make(map[uint]int, len(items))
	for _, it := range items {
		productIDs = append(productIDs, it.ProductID)
		qtyByProduct[it.ProductID] += it.Quantity
	}

	// Traigo stock disponible por (product_id, branch_id).
	var stocks []models.ProductBranchStock
	if err := r.db.
		Where("product_id IN ?", productIDs).
		Find(&stocks).Error; err != nil {
		return nil, err
	}

	coverage := make(map[uint]int)
	for _, s := range stocks {
		needed, ok := qtyByProduct[s.ProductID]
		if !ok {
			continue
		}
		if s.Stock-s.Reserved >= needed {
			coverage[s.BranchID]++
		}
	}

	needed := len(qtyByProduct)
	var result []uint
	for branchID, c := range coverage {
		if c == needed {
			result = append(result, branchID)
		}
	}
	_ = rows
	return result, nil
}

// OutOfStockItemsForBranch lista los product_ids que NO tienen stock suficiente
// en la sucursal dada.
func (r *productBranchStockRepository) OutOfStockItemsForBranch(branchID uint, items []models.QuoteItem) ([]uint, error) {
	if len(items) == 0 {
		return nil, nil
	}
	qtyByProduct := make(map[uint]int, len(items))
	productIDs := make([]uint, 0, len(items))
	for _, it := range items {
		qtyByProduct[it.ProductID] += it.Quantity
		productIDs = append(productIDs, it.ProductID)
	}

	var stocks []models.ProductBranchStock
	if err := r.db.
		Where("branch_id = ? AND product_id IN ?", branchID, productIDs).
		Find(&stocks).Error; err != nil {
		return nil, err
	}

	available := make(map[uint]int, len(stocks))
	for _, s := range stocks {
		available[s.ProductID] = s.Stock - s.Reserved
	}

	var missing []uint
	for pid, need := range qtyByProduct {
		if available[pid] < need {
			missing = append(missing, pid)
		}
	}
	return missing, nil
}
