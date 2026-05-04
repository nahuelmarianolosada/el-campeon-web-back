package report

import (
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type ReportService interface {
	GetOrdersReport() ([]models.OrderReportItem, error)
	GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error)
	GetDailyRevenueReport() ([]models.DailyRevenue, error)
}

type reportService struct {
	reportRepo repositories.ReportRepository
}

func NewReportService(reportRepo repositories.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) GetOrdersReport() ([]models.OrderReportItem, error) {
	return s.reportRepo.GetOrdersReport()
}

func (s *reportService) GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.reportRepo.GetLowStockProductsReport(limit)
}

func (s *reportService) GetDailyRevenueReport() ([]models.DailyRevenue, error) {
	return s.reportRepo.GetDailyRevenueReport()
}
