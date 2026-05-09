package payment

import (
	"context"
	"errors"
	"testing"

	preferenceMp "github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	orderStatus "github.com/nahuelmarianolosada/el-campeon-web/internal/services/order/status"
	paymentStatus "github.com/nahuelmarianolosada/el-campeon-web/internal/services/payment/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks

type MockMercadopagoClient struct {
	mock.Mock
}

func (m *MockMercadopagoClient) CreatePreference(ctx context.Context, req preferenceMp.Request) (*preferenceMp.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*preferenceMp.Response), args.Error(1)
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(id uint) (*models.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByOrderID(orderID uint) (*models.Payment, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByUserID(userID uint, limit, offset int) ([]models.Payment, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) ListAll(limit, offset int) ([]models.Payment, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(paymentID uint, status string) error {
	args := m.Called(paymentID, status)
	return args.Error(0)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(id uint) (*models.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByOrderNumber(orderNumber string) (*models.Order, error) {
	args := m.Called(orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUserID(userID uint, limit, offset int) ([]models.Order, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrderRepository) AddItem(orderID uint, item *models.OrderItem) error {
	args := m.Called(orderID, item)
	return args.Error(0)
}

func (m *MockOrderRepository) FindAll(limit, offset int) ([]models.Order, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(orderID uint, status string) error {
	args := m.Called(orderID, status)
	return args.Error(0)
}

// Tests

func TestCreatePayment(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	orderRepo := new(MockOrderRepository)
	mpClient := new(MockMercadopagoClient)
	cfg := &config.Config{}
	service := NewPaymentServiceWithClient(paymentRepo, orderRepo, cfg, mpClient)
	ctx := context.Background()

	req := &models.CreatePaymentRequest{
		OrderID:       1,
		Amount:        100.0,
		PaymentMethod: "MP_CARD",
	}

	order := &models.Order{
		ID:     1,
		UserID: 1,
		Total:  100.0,
		Status: orderStatus.Pending,
		User: &models.User{
			ID:        1,
			FirstName: "Test User",
			LastName:  "Test Last Name",
		},
		Items: []models.OrderItem{
			{
				ID:       1,
				Quantity: 1,
				Price:    100.0,
				Product: &models.Product{
					Name: "Test Product",
				},
			},
		},
	}

	t.Run("SuccessMPCard", func(t *testing.T) {
		orderRepo.On("FindByID", uint(1)).Return(order, nil).Once()

		// Setup the mock Mercadopago client to return a successful preference response
		mockPaymentResponse := &preferenceMp.Response{
			ID: "123456",
		}
		mpClient.On("CreatePreference", mock.Anything, mock.Anything).Return(mockPaymentResponse, nil).Once()

		paymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		paymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		orderRepo.On("Update", mock.AnythingOfType("*models.Order")).Return(nil).Once()

		resp, err := service.CreatePayment(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.OrderID, resp.OrderID)
		assert.Equal(t, order.UserID, resp.UserID)
		assert.Equal(t, req.Amount, resp.Amount)
		assert.Equal(t, paymentStatus.Pending, resp.Status)
		assert.Equal(t, "123456", resp.MercadopagoPreferenceID)
		assert.Equal(t, "MP_CARD", resp.PaymentMethod)
		paymentRepo.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
		mpClient.AssertExpectations(t)
	})

	t.Run("SuccessCash", func(t *testing.T) {
		freshRepo := new(MockPaymentRepository)
		freshOrderRepo := new(MockOrderRepository)
		freshService := NewPaymentService(freshRepo, freshOrderRepo, nil)

		reqCash := &models.CreatePaymentRequest{
			OrderID:       1,
			Amount:        100.0,
			PaymentMethod: "CASH",
		}

		freshOrderRepo.On("FindByID", uint(1)).Return(order, nil).Once()
		freshRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		freshRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil).Once()

		resp, err := freshService.CreatePayment(ctx, reqCash)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "CASH", resp.PaymentMethod)
		assert.Equal(t, paymentStatus.Pending, resp.Status)
		freshRepo.AssertExpectations(t)
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		orderRepo.On("FindByID", uint(1)).Return(nil, errors.New("not found")).Once()

		resp, err := service.CreatePayment(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "error finding order")
		orderRepo.AssertExpectations(t)
	})

	t.Run("OrderCancelled", func(t *testing.T) {
		cancelledOrder := &models.Order{ID: 1, Status: "CANCELLED"}
		orderRepo.On("FindByID", uint(1)).Return(cancelledOrder, nil).Once()

		resp, err := service.CreatePayment(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "cannot create payment for cancelled order", err.Error())
		orderRepo.AssertExpectations(t)
	})

	t.Run("AmountMismatch", func(t *testing.T) {
		orderRepo.On("FindByID", uint(1)).Return(order, nil).Once()
		reqMismatch := &models.CreatePaymentRequest{OrderID: 1, Amount: 50.0, PaymentMethod: "MP_CARD"}

		resp, err := service.CreatePayment(ctx, reqMismatch)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "payment amount does not match order total")
		orderRepo.AssertExpectations(t)
	})

	t.Run("RepoCreateError", func(t *testing.T) {
		orderRepo.On("FindByID", uint(1)).Return(order, nil).Once()
		paymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(errors.New("db error")).Once()

		resp, err := service.CreatePayment(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "error creating payment")
		paymentRepo.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
	})
}

func TestGetPaymentByID(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	service := NewPaymentService(paymentRepo, nil, nil)

	paymentData := &models.Payment{ID: 1, OrderID: 10}

	t.Run("Success", func(t *testing.T) {
		paymentRepo.On("FindByID", uint(1)).Return(paymentData, nil).Once()

		resp, err := service.GetPaymentByID(1)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), resp.ID)
		paymentRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		paymentRepo.On("FindByID", uint(2)).Return(nil, errors.New("not found")).Once()

		resp, err := service.GetPaymentByID(2)

		assert.Error(t, err)
		assert.Nil(t, resp)
		paymentRepo.AssertExpectations(t)
	})
}

func TestGetPaymentsByUserID(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	service := NewPaymentService(paymentRepo, nil, nil)

	payments := []models.Payment{
		{ID: 1, UserID: 100},
		{ID: 2, UserID: 100},
	}

	t.Run("Success", func(t *testing.T) {
		paymentRepo.On("FindByUserID", uint(100), 10, 0).Return(payments, nil).Once()

		resp, err := service.GetPaymentsByUserID(100, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, uint(1), resp[0].ID)
		paymentRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		paymentRepo.On("FindByUserID", uint(100), 10, 0).Return(nil, errors.New("db error")).Once()

		resp, err := service.GetPaymentsByUserID(100, 10, 0)

		assert.Error(t, err)
		assert.Nil(t, resp)
		paymentRepo.AssertExpectations(t)
	})
}

func TestGetPaymentByOrderID(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	service := NewPaymentService(paymentRepo, nil, nil)

	paymentData := &models.Payment{ID: 1, OrderID: 10}

	t.Run("Success", func(t *testing.T) {
		paymentRepo.On("FindByOrderID", uint(10)).Return(paymentData, nil).Once()

		resp, err := service.GetPaymentByOrderID(10)

		assert.NoError(t, err)
		assert.Equal(t, uint(10), resp.OrderID)
		paymentRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		paymentRepo.On("FindByOrderID", uint(10)).Return(nil, errors.New("not found")).Once()

		resp, err := service.GetPaymentByOrderID(10)

		assert.Error(t, err)
		assert.Nil(t, resp)
		paymentRepo.AssertExpectations(t)
	})
}

func TestUpdatePaymentStatus(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	orderRepo := new(MockOrderRepository)
	service := NewPaymentService(paymentRepo, orderRepo, nil)

	paymentData := &models.Payment{ID: 1, OrderID: 10, Status: "PENDING"}

	t.Run("ApproveSuccess", func(t *testing.T) {
		paymentRepo.On("FindByID", uint(1)).Return(paymentData, nil).Once()
		orderRepo.On("UpdateStatus", uint(10), "CONFIRMED").Return(nil).Once()
		paymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*models.Payment)
			assert.Equal(t, "APPROVED", p.Status)
			assert.NotNil(t, p.ApprovedAt)
		}).Once()

		resp, err := service.UpdatePaymentStatus(1, "APPROVED")

		assert.NoError(t, err)
		assert.Equal(t, "APPROVED", resp.Status)
		paymentRepo.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
	})

	t.Run("RejectSuccess", func(t *testing.T) {
		freshPayment := &models.Payment{ID: 1, OrderID: 10, Status: "PENDING"}
		paymentRepo.On("FindByID", uint(1)).Return(freshPayment, nil).Once()
		paymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil).Run(func(args mock.Arguments) {
			p := args.Get(0).(*models.Payment)
			assert.Equal(t, "REJECTED", p.Status)
		}).Once()

		resp, err := service.UpdatePaymentStatus(1, "REJECTED")

		assert.NoError(t, err)
		assert.Equal(t, "REJECTED", resp.Status)
		paymentRepo.AssertExpectations(t)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		freshRepo := new(MockPaymentRepository)
		freshOrderRepo := new(MockOrderRepository)
		freshService := NewPaymentService(freshRepo, freshOrderRepo, nil)
		invalidPayment := &models.Payment{ID: 1, OrderID: 10, Status: "PENDING"}
		freshRepo.On("FindByID", uint(1)).Return(invalidPayment, nil).Once()

		resp, err := freshService.UpdatePaymentStatus(1, "INVALID")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid payment status transition")
		freshRepo.AssertExpectations(t)
	})

	t.Run("PaymentNotFound", func(t *testing.T) {
		freshRepo := new(MockPaymentRepository)
		freshOrderRepo := new(MockOrderRepository)
		freshService := NewPaymentService(freshRepo, freshOrderRepo, nil)
		freshRepo.On("FindByID", uint(1)).Return(nil, errors.New("not found")).Once()

		resp, err := freshService.UpdatePaymentStatus(1, "APPROVED")

		assert.Error(t, err)
		assert.Nil(t, resp)
		freshRepo.AssertExpectations(t)
	})

	t.Run("OrderUpdateError", func(t *testing.T) {
		freshRepo := new(MockPaymentRepository)
		freshOrderRepo := new(MockOrderRepository)
		freshService := NewPaymentService(freshRepo, freshOrderRepo, nil)
		errorPayment := &models.Payment{ID: 1, OrderID: 10, Status: "PENDING"}
		freshRepo.On("FindByID", uint(1)).Return(errorPayment, nil).Once()
		freshOrderRepo.On("UpdateStatus", uint(10), "CONFIRMED").Return(errors.New("db error")).Once()

		resp, err := freshService.UpdatePaymentStatus(1, "APPROVED")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "error updating order status")
		freshOrderRepo.AssertExpectations(t)
	})

	t.Run("PaymentUpdateError", func(t *testing.T) {
		freshRepo := new(MockPaymentRepository)
		freshOrderRepo := new(MockOrderRepository)
		freshService := NewPaymentService(freshRepo, freshOrderRepo, nil)
		errorPayment := &models.Payment{ID: 1, OrderID: 10, Status: "PENDING"}
		freshRepo.On("FindByID", uint(1)).Return(errorPayment, nil).Once()
		freshRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(errors.New("db error")).Once()

		resp, err := freshService.UpdatePaymentStatus(1, "REJECTED")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "error updating payment")
		freshRepo.AssertExpectations(t)
	})
}

func TestListAllPayments(t *testing.T) {
	paymentRepo := new(MockPaymentRepository)
	service := NewPaymentService(paymentRepo, nil, nil)

	payments := []models.Payment{
		{ID: 1},
		{ID: 2},
	}

	t.Run("Success", func(t *testing.T) {
		paymentRepo.On("ListAll", 10, 0).Return(payments, nil).Once()

		resp, err := service.ListAllPayments(10, 0)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		paymentRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		paymentRepo.On("ListAll", 10, 0).Return(nil, errors.New("db error")).Once()

		resp, err := service.ListAllPayments(10, 0)

		assert.Error(t, err)
		assert.Nil(t, resp)
		paymentRepo.AssertExpectations(t)
	})
}

func TestProcessMercadopagoWebhook(t *testing.T) {
	service := NewPaymentService(nil, nil, nil)

	t.Run("NotPaymentType", func(t *testing.T) {
		webhook := &models.MercadopagoWebhookRequest{
			Type: "other",
		}
		err := service.ProcessMercadopagoWebhook(webhook)
		assert.NoError(t, err)
	})

	t.Run("PaymentType", func(t *testing.T) {
		webhook := &models.MercadopagoWebhookRequest{
			Type:   "payment",
			Action: "payment.created",
			Data: struct {
				ID string `json:"id"`
			}{ID: "123"},
		}
		err := service.ProcessMercadopagoWebhook(webhook)
		assert.NoError(t, err)
	})
}
