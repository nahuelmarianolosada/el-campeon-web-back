package report

import (
	"log"

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
	log.Printf("[reportService.GetOrdersReport] INFO: Retrieving orders report")

	report, err := s.reportRepo.GetOrdersReport()
	if err != nil {
		log.Printf("[reportService.GetOrdersReport] ERROR: Failed to retrieve report: %v", err)
		return nil, err
	}
	log.Printf("[reportService.GetOrdersReport] INFO: Orders report retrieved - itemCount=%d", len(report))
	return report, nil
}

func (s *reportService) GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error) {
	log.Printf("[reportService.GetLowStockProductsReport] INFO: Retrieving low stock products report - limit=%d", limit)

	if limit <= 0 {
		log.Printf("[reportService.GetLowStockProductsReport] WARNING: Invalid limit provided - limit=%d, using default limit=10", limit)
		limit = 10
	}

	report, err := s.reportRepo.GetLowStockProductsReport(limit)
	if err != nil {
		log.Printf("[reportService.GetLowStockProductsReport] ERROR: Failed to retrieve report: %v", err)
		return nil, err
	}
	log.Printf("[reportService.GetLowStockProductsReport] INFO: Low stock products report retrieved - productCount=%d", len(report))
	return report, nil
}

func (s *reportService) GetDailyRevenueReport() ([]models.DailyRevenue, error) {
	log.Printf("[reportService.GetDailyRevenueReport] INFO: Retrieving daily revenue report")

	report, err := s.reportRepo.GetDailyRevenueReport()
	if err != nil {
		log.Printf("[reportService.GetDailyRevenueReport] ERROR: Failed to retrieve report: %v", err)
		return nil, err
	}
	log.Printf("[reportService.GetDailyRevenueReport] INFO: Daily revenue report retrieved - dayCount=%d", len(report))
	return report, nil
}
