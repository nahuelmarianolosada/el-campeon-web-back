package payment

import (
	"fmt"
	"time"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
)

type PaymentService interface {
	CreatePayment(req *models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentByID(id uint) (*models.PaymentResponse, error)
	GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error)
	GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error)
	UpdatePaymentStatus(paymentID uint, status string) (*models.PaymentResponse, error)
	ProcessMercadopagoWebhook(webhook *models.MercadopagoWebhookRequest) error
	ListAllPayments(limit, offset int) ([]models.PaymentResponse, error)
}

type paymentService struct {
	paymentRepo repositories.PaymentRepository
	orderRepo   repositories.OrderRepository
	config      *config.Config
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cfg *config.Config,
) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		config:      cfg,
	}
}

func (s *paymentService) CreatePayment(req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// Obtener la orden
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	// Validar que la orden no esté cancelada
	if order.Status == "CANCELLED" {
		return nil, fmt.Errorf("cannot create payment for cancelled order")
	}

	// Verificar que el monto coincide
	if req.Amount != order.Total {
		return nil, fmt.Errorf("payment amount does not match order total. expected: %.2f, got: %.2f", order.Total, req.Amount)
	}

	// Crear pago
	transactionID := s.generateTransactionID()
	payment := &models.Payment{
		TransactionID:           transactionID,
		OrderID:                 req.OrderID,
		UserID:                  order.UserID,
		Amount:                  req.Amount,
		Currency:                "ARS",
		Status:                  "PENDING",
		PaymentMethod:           "MERCADOPAGO",
		MercadopagoPreferenceID: "", // Se generará al integrar con MP SDK
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) GetPaymentByID(id uint) (*models.PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding payment: %w", err)
	}

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding payments: %w", err)
	}

	var responses []models.PaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.toPaymentResponse(&payment))
	}

	return responses, nil
}

func (s *paymentService) GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("error finding payment for order: %w", err)
	}

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) UpdatePaymentStatus(paymentID uint, status string) (*models.PaymentResponse, error) {
	// Validar estado
	validStatuses := map[string]bool{
		"PENDING":   true,
		"APPROVED":  true,
		"REJECTED":  true,
		"CANCELLED": true,
		"REFUNDED":  true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid payment status: %s", status)
	}

	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("error finding payment: %w", err)
	}

	payment.Status = status

	// Si el pago fue aprobado, actualizar estado de orden
	if status == "APPROVED" {
		now := time.Now()
		payment.ApprovedAt = &now

		// Actualizar orden a CONFIRMED
		if err := s.orderRepo.UpdateStatus(payment.OrderID, "CONFIRMED"); err != nil {
			return nil, fmt.Errorf("error updating order status: %w", err)
		}
	}

	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("error updating payment: %w", err)
	}

	return s.toPaymentResponse(payment), nil
}

// ProcessMercadopagoWebhook procesa webhooks de MercadoPago
// En producción, aquí se verificaría la firma del webhook y se integraría con MP SDK
func (s *paymentService) ProcessMercadopagoWebhook(webhook *models.MercadopagoWebhookRequest) error {
	if webhook.Type != "payment" {
		return nil // Ignorar otros tipos de eventos
	}

	// En una implementación real:
	// 1. Verificar la firma del webhook
	// 2. Consultar el estado del pago en MercadoPago API
	// 3. Actualizar el perfil del pago con los datos de MP
	// 4. Actualizar el estado basado en la respuesta de MP

	// Por ahora, esto es un placeholder
	fmt.Printf("Received webhook for payment %s with action %s\n", webhook.Data.ID, webhook.Action)

	return nil
}

func (s *paymentService) ListAllPayments(limit, offset int) ([]models.PaymentResponse, error) {
	payments, err := s.paymentRepo.ListAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing payments: %w", err)
	}

	var responses []models.PaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.toPaymentResponse(&payment))
	}

	return responses, nil
}

// Helper functions

func (s *paymentService) generateTransactionID() string {
	return fmt.Sprintf("TXN-%d", time.Now().UnixNano())
}

func (s *paymentService) toPaymentResponse(payment *models.Payment) *models.PaymentResponse {
	return &models.PaymentResponse{
		ID:                      payment.ID,
		TransactionID:           payment.TransactionID,
		OrderID:                 payment.OrderID,
		UserID:                  payment.UserID,
		Amount:                  payment.Amount,
		Currency:                payment.Currency,
		Status:                  payment.Status,
		PaymentMethod:           payment.PaymentMethod,
		MercadopagoPreferenceID: payment.MercadopagoPreferenceID,
		ApprovedAt:              payment.ApprovedAt,
		RejectedReason:          payment.RejectedReason,
		CreatedAt:               payment.CreatedAt,
		UpdatedAt:               payment.UpdatedAt,
	}
}
