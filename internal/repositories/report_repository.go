package repositories

import (
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type ReportRepository interface {
	GetOrdersReport() ([]models.OrderReportItem, error)
	GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error)
	GetDailyRevenueReport() ([]models.DailyRevenue, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetOrdersReport() ([]models.OrderReportItem, error) {
	var report []models.OrderReportItem
	query := `
		SELECT 
		  o.id,
		  o.order_number,
		  u.email,
		  GROUP_CONCAT(p.name SEPARATOR ',') as productos,
		  SUM(oi.quantity) as cantidad_items,
		  o.total,
		  o.status,
		  o.created_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		WHERE o.deleted_at IS NULL
		GROUP BY o.id
		ORDER BY o.created_at DESC
	`
	err := r.db.Raw(query).Scan(&report).Error
	return report, err
}

func (r *reportRepository) GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error) {
	var report []models.LowStockProduct
	query := `
		SELECT 
		  id,
		  sku,
		  name,
		  stock,
		  category
		FROM products
		WHERE stock < ? AND is_active = TRUE
		ORDER BY stock ASC
	`
	err := r.db.Raw(query, limit).Scan(&report).Error
	return report, err
}

func (r *reportRepository) GetDailyRevenueReport() ([]models.DailyRevenue, error) {
	var report []models.DailyRevenue
	query := `
		SELECT 
		  DATE(created_at) as fecha,
		  COUNT(*) as cantidad_ordenes,
		  SUM(total) as ingresos
		FROM orders
		WHERE deleted_at IS NULL AND status IN ('CONFIRMED', 'SHIPPED', 'DELIVERED')
		GROUP BY DATE(created_at)
		ORDER BY fecha DESC
	`
	err := r.db.Raw(query).Scan(&report).Error
	return report, err
}
