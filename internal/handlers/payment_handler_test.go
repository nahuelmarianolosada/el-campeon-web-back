package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPaymentService struct {
	mock.Mock
}

func (m *mockPaymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) CreateGuestPayment(ctx context.Context, req *models.CreateGuestPaymentRequest) (*models.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) GetPaymentByID(id uint) (*models.PaymentResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) UpdatePaymentStatus(id uint, status string) (*models.PaymentResponse, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) CancelPayment(ctx context.Context, paymentID uint) (*models.PaymentResponse, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) ProcessMercadopagoWebhook(ctx context.Context, webhook *models.MercadopagoWebhookRequest, xSignature string, xRequestId string) error {
	args := m.Called(ctx, webhook, xSignature, xRequestId)
	return args.Error(0)
}

func (m *mockPaymentService) ListAllPayments(limit, offset int) ([]models.PaymentResponse, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PaymentResponse), args.Error(1)
}

func newPaymentRouter(svc *mockPaymentService) (*gin.Engine, *PaymentHandler) {
	h := NewPaymentHandler(svc)
	r := gin.New()
	return r, h
}

func TestCreatePayment_Unauthenticated_Returns401(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/payments", h.CreatePayment)

	req := httptest.NewRequest("POST", "/payments",
		bytes.NewBufferString(`{"order_id":1,"amount":100,"payment_method":"CASH"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreatePayment_InvalidJSON_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/payments", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CreatePayment(c) })

	req := httptest.NewRequest("POST", "/payments", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePayment_Success(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CreatePayment", mock.Anything, mock.Anything).Return(
		&models.PaymentResponse{ID: 99, Status: "PENDING"}, nil,
	)
	r, h := newPaymentRouter(svc)
	r.POST("/payments", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CreatePayment(c) })

	req := httptest.NewRequest("POST", "/payments",
		bytes.NewBufferString(`{"order_id":1,"amount":100,"payment_method":"CASH"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":99`)
}

func TestCreatePayment_ServiceError_Returns500(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CreatePayment", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
	r, h := newPaymentRouter(svc)
	r.POST("/payments", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CreatePayment(c) })

	req := httptest.NewRequest("POST", "/payments",
		bytes.NewBufferString(`{"order_id":1,"amount":100,"payment_method":"CASH"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateGuestPayment_OverridesEmailFromContext(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CreateGuestPayment", mock.Anything, mock.MatchedBy(func(r *models.CreateGuestPaymentRequest) bool {
		// El email del contexto (autenticado por X-Guest-Token) debe pisar el del body.
		return r.Email == "ctx@example.com"
	})).Return(&models.PaymentResponse{ID: 1}, nil)

	r, h := newPaymentRouter(svc)
	r.POST("/payments/guest", func(c *gin.Context) {
		c.Set("guest_email", "ctx@example.com")
		h.CreateGuestPayment(c)
	})

	req := httptest.NewRequest("POST", "/payments/guest", bytes.NewBufferString(
		`{"order_id":1,"email":"body@example.com","amount":100,"payment_method":"CASH"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCreateGuestPayment_NoContext_UsesBodyEmail(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CreateGuestPayment", mock.Anything, mock.MatchedBy(func(r *models.CreateGuestPaymentRequest) bool {
		return r.Email == "body@example.com"
	})).Return(&models.PaymentResponse{ID: 1}, nil)

	r, h := newPaymentRouter(svc)
	r.POST("/payments/guest", h.CreateGuestPayment)

	req := httptest.NewRequest("POST", "/payments/guest", bytes.NewBufferString(
		`{"order_id":1,"email":"body@example.com","amount":100,"payment_method":"CASH"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestGetPayment_InvalidID_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/:id", h.GetPayment)

	req := httptest.NewRequest("GET", "/payments/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPayment_NotFound_Returns404(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("GetPaymentByID", uint(5)).Return(nil, errors.New("not found"))
	r, h := newPaymentRouter(svc)
	r.GET("/payments/:id", h.GetPayment)

	req := httptest.NewRequest("GET", "/payments/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPayment_Found_Returns200(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("GetPaymentByID", uint(5)).Return(&models.PaymentResponse{ID: 5, Status: "APPROVED"}, nil)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/:id", h.GetPayment)

	req := httptest.NewRequest("GET", "/payments/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"APPROVED"`)
}

func TestGetMyPayments_Unauthenticated_Returns401(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/my", h.GetMyPayments)

	req := httptest.NewRequest("GET", "/payments/my", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMyPayments_DefaultsLimitAndOffset(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("GetPaymentsByUserID", uint(1), 20, 0).Return([]models.PaymentResponse{}, nil)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/my", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetMyPayments(c) })

	req := httptest.NewRequest("GET", "/payments/my", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestGetMyPayments_AppliesQueryParams(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("GetPaymentsByUserID", uint(1), 5, 10).Return([]models.PaymentResponse{}, nil)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/my", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetMyPayments(c) })

	req := httptest.NewRequest("GET", "/payments/my?limit=5&offset=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestGetMyPayments_IgnoresInvalidQueryParams(t *testing.T) {
	svc := new(mockPaymentService)
	// limit inválido y offset negativo deben caer al default.
	svc.On("GetPaymentsByUserID", uint(1), 20, 0).Return([]models.PaymentResponse{}, nil)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/my", func(c *gin.Context) { c.Set("user_id", uint(1)); h.GetMyPayments(c) })

	req := httptest.NewRequest("GET", "/payments/my?limit=abc&offset=-3", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestGetPaymentByOrderID_InvalidID_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.GET("/payments/order/:orderId", h.GetPaymentByOrderID)

	req := httptest.NewRequest("GET", "/payments/order/xyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePaymentStatus_Success(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("UpdatePaymentStatus", uint(7), "APPROVED").Return(
		&models.PaymentResponse{ID: 7, Status: "APPROVED"}, nil,
	)
	r, h := newPaymentRouter(svc)
	r.PUT("/payments/:id/status", h.UpdatePaymentStatus)

	req := httptest.NewRequest("PUT", "/payments/7/status", bytes.NewBufferString(`{"status":"APPROVED"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdatePaymentStatus_InvalidID_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.PUT("/payments/:id/status", h.UpdatePaymentStatus)

	req := httptest.NewRequest("PUT", "/payments/abc/status", bytes.NewBufferString(`{"status":"APPROVED"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePaymentStatus_MissingStatusInBody_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.PUT("/payments/:id/status", h.UpdatePaymentStatus)

	req := httptest.NewRequest("PUT", "/payments/7/status", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelPayment_Unauthenticated_Returns401(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/payments/:id/cancel", h.CancelPayment)

	req := httptest.NewRequest("POST", "/payments/1/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCancelPayment_InvalidID_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/payments/:id/cancel", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CancelPayment(c) })

	req := httptest.NewRequest("POST", "/payments/abc/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelPayment_Success(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CancelPayment", mock.Anything, uint(1)).Return(
		&models.PaymentResponse{ID: 1, Status: "CANCELLED"}, nil,
	)
	r, h := newPaymentRouter(svc)
	r.POST("/payments/:id/cancel", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CancelPayment(c) })

	req := httptest.NewRequest("POST", "/payments/1/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"CANCELLED"`)
	svc.AssertExpectations(t)
}

func TestCancelPayment_ServiceError_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("CancelPayment", mock.Anything, uint(1)).Return(nil, errors.New("payment cannot be cancelled from status REJECTED"))
	r, h := newPaymentRouter(svc)
	r.POST("/payments/:id/cancel", func(c *gin.Context) { c.Set("user_id", uint(1)); h.CancelPayment(c) })

	req := httptest.NewRequest("POST", "/payments/1/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestMercadopagoWebhook_MissingSignature_Returns401(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/webhook", h.MercadopagoWebhook)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(`{"type":"payment","data":{"id":"123"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing X-Signature")
}

func TestMercadopagoWebhook_InvalidJSON_Returns400(t *testing.T) {
	svc := new(mockPaymentService)
	r, h := newPaymentRouter(svc)
	r.POST("/webhook", h.MercadopagoWebhook)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMercadopagoWebhook_Success(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("ProcessMercadopagoWebhook", mock.Anything, mock.Anything, "sig", "rid").Return(nil)
	r, h := newPaymentRouter(svc)
	r.POST("/webhook", h.MercadopagoWebhook)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(`{"type":"payment","data":{"id":"123"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "sig")
	req.Header.Set("X-Request-Id", "rid")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "webhook processed")
	svc.AssertExpectations(t)
}

func TestMercadopagoWebhook_ServiceError_Returns500(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("ProcessMercadopagoWebhook", mock.Anything, mock.Anything, "sig", "rid").Return(errors.New("kaboom"))
	r, h := newPaymentRouter(svc)
	r.POST("/webhook", h.MercadopagoWebhook)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(`{"type":"payment","data":{"id":"123"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "sig")
	req.Header.Set("X-Request-Id", "rid")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListAllPayments_DefaultsLimitOffset(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("ListAllPayments", 20, 0).Return([]models.PaymentResponse{}, nil)
	r, h := newPaymentRouter(svc)
	r.GET("/payments", h.ListAllPayments)

	req := httptest.NewRequest("GET", "/payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestListAllPayments_ServiceError_Returns500(t *testing.T) {
	svc := new(mockPaymentService)
	svc.On("ListAllPayments", 20, 0).Return(nil, errors.New("db down"))
	r, h := newPaymentRouter(svc)
	r.GET("/payments", h.ListAllPayments)

	req := httptest.NewRequest("GET", "/payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
