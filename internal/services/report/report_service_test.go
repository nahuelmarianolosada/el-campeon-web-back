package report

import (
	"errors"
	"testing"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
)

// MockReportRepository implements repositories.ReportRepository
type MockReportRepository struct {
	GetOrdersReportFunc           func() ([]models.OrderReportItem, error)
	GetLowStockProductsReportFunc func(limit int) ([]models.LowStockProduct, error)
	GetDailyRevenueReportFunc     func() ([]models.DailyRevenue, error)
}

func (m *MockReportRepository) GetOrdersReport() ([]models.OrderReportItem, error) {
	if m.GetOrdersReportFunc != nil {
		return m.GetOrdersReportFunc()
	}
	return nil, nil
}

func (m *MockReportRepository) GetLowStockProductsReport(limit int) ([]models.LowStockProduct, error) {
	if m.GetLowStockProductsReportFunc != nil {
		return m.GetLowStockProductsReportFunc(limit)
	}
	return nil, nil
}

func (m *MockReportRepository) GetDailyRevenueReport() ([]models.DailyRevenue, error) {
	if m.GetDailyRevenueReportFunc != nil {
		return m.GetDailyRevenueReportFunc()
	}
	return nil, nil
}

func TestGetOrdersReport(t *testing.T) {
	mockRepo := &MockReportRepository{}
	service := NewReportService(mockRepo)

	expectedReport := []models.OrderReportItem{
		{ID: 1, OrderNumber: "ORD-001", Email: "user@example.com"},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.GetOrdersReportFunc = func() ([]models.OrderReportItem, error) {
			return expectedReport, nil
		}

		result, err := service.GetOrdersReport()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 1 || result[0].OrderNumber != "ORD-001" {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.GetOrdersReportFunc = func() ([]models.OrderReportItem, error) {
			return nil, errors.New("db error")
		}

		_, err := service.GetOrdersReport()
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestGetLowStockProductsReport(t *testing.T) {
	mockRepo := &MockReportRepository{}
	service := NewReportService(mockRepo)

	t.Run("Success with default limit", func(t *testing.T) {
		capturedLimit := 0
		mockRepo.GetLowStockProductsReportFunc = func(limit int) ([]models.LowStockProduct, error) {
			capturedLimit = limit
			return []models.LowStockProduct{{ID: 1, Name: "Product 1"}}, nil
		}

		_, err := service.GetLowStockProductsReport(0)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if capturedLimit != 10 {
			t.Errorf("Expected default limit 10, got %d", capturedLimit)
		}
	})

	t.Run("Success with custom limit", func(t *testing.T) {
		capturedLimit := 0
		mockRepo.GetLowStockProductsReportFunc = func(limit int) ([]models.LowStockProduct, error) {
			capturedLimit = limit
			return []models.LowStockProduct{{ID: 1, Name: "Product 1"}}, nil
		}

		_, err := service.GetLowStockProductsReport(5)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if capturedLimit != 5 {
			t.Errorf("Expected limit 5, got %d", capturedLimit)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.GetLowStockProductsReportFunc = func(limit int) ([]models.LowStockProduct, error) {
			return nil, errors.New("db error")
		}

		_, err := service.GetLowStockProductsReport(10)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestGetDailyRevenueReport(t *testing.T) {
	mockRepo := &MockReportRepository{}
	service := NewReportService(mockRepo)

	expectedReport := []models.DailyRevenue{
		{Fecha: "2023-01-01", CantidadOrdenes: 5, Ingresos: 500.0},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.GetDailyRevenueReportFunc = func() ([]models.DailyRevenue, error) {
			return expectedReport, nil
		}

		result, err := service.GetDailyRevenueReport()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 1 || result[0].Ingresos != 500.0 {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.GetDailyRevenueReportFunc = func() ([]models.DailyRevenue, error) {
			return nil, errors.New("db error")
		}

		_, err := service.GetDailyRevenueReport()
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
