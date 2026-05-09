package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mercadopago/sdk-go/pkg/config"
	preferenceMp "github.com/mercadopago/sdk-go/pkg/preference"
	internalConfig "github.com/nahuelmarianolosada/el-campeon-web/internal/config"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/repositories"
	orderStatus "github.com/nahuelmarianolosada/el-campeon-web/internal/services/order/status"
	paymentStatus "github.com/nahuelmarianolosada/el-campeon-web/internal/services/payment/status"
	"gorm.io/datatypes"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentByID(id uint) (*models.PaymentResponse, error)
	GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error)
	GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error)
	UpdatePaymentStatus(paymentID uint, status string) (*models.PaymentResponse, error)
	ProcessMercadopagoWebhook(webhook *models.MercadopagoWebhookRequest) error
	ListAllPayments(limit, offset int) ([]models.PaymentResponse, error)
}

type paymentService struct {
	paymentRepo       repositories.PaymentRepository
	orderRepo         repositories.OrderRepository
	config            *internalConfig.Config
	mercadopagoClient MercadopagoClient
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cfg *internalConfig.Config,
) PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		orderRepo:         orderRepo,
		config:            cfg,
		mercadopagoClient: nil, // Will be set via SetMercadopagoClient or use default
	}
}

func NewPaymentServiceWithClient(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cfg *internalConfig.Config,
	client MercadopagoClient,
) PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		orderRepo:         orderRepo,
		config:            cfg,
		mercadopagoClient: client,
	}
}

func (s *paymentService) SetMercadopagoClient(client MercadopagoClient) {
	s.mercadopagoClient = client
}

func (s *paymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// Obtener la orden
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	// Validar que la orden no esté cancelada
	if order.Status == orderStatus.Cancelled {
		return nil, fmt.Errorf("cannot create payment for cancelled order")
	}

	// Verificar que el monto coincide
	if req.Amount != order.Total {
		return nil, fmt.Errorf("payment amount does not match order total. expected: %.2f, got: %.2f", order.Total, req.Amount)
	}

	// Crear pago
	transactionID := s.generateTransactionID()
	payment := &models.Payment{
		TransactionID: transactionID,
		OrderID:       req.OrderID,
		UserID:        order.UserID,
		Amount:        req.Amount,
		Currency:      "ARS",
		Status:        paymentStatus.Pending,
		PaymentMethod: req.PaymentMethod,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	// Para pagos en efectivo, marcar como pendiente de confirmación
	if req.PaymentMethod == "CASH" {
		payment.Status = paymentStatus.Pending
		// No procesamos con MercadoPago para pagos en efectivo
		if err := s.paymentRepo.Update(payment); err != nil {
			return nil, fmt.Errorf("error updating payment: %w", err)
		}
		return s.toPaymentResponse(payment), nil
	}

	// Para pagos con MercadoPago, crear preference
	executedPayment, err := s.ExecutePayment(ctx, *order, req.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("error executing payment: %w", err)
	}

	executedPaymentByte, err := json.Marshal(executedPayment)
	if err != nil {
		return nil, fmt.Errorf("error marshaling executed payment: %w", err)
	}

	payment.MercadopagoPreferenceID = executedPayment.ID
	payment.MercadopagoData = datatypes.JSONMap{
		"preference": string(executedPaymentByte),
	}
	payment.Status = paymentStatus.Pending // En Mercado Pago, el pago sigue siendo PENDING hasta confirmación de webhook

	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("error updating payment with MP data: %w", err)
	}

	// Actualizar orden a CONFIRMED
	order.Status = orderStatus.Confirmed
	order.UpdatedAt = time.Now()
	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("error updating order status: %w", err)
	}

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) ExecutePayment(ctx context.Context, order models.Order, paymentMethod string) (*preferenceMp.Response, error) {
	// If no client is set, create a default one
	if s.mercadopagoClient == nil {
		cfg, err := config.New(s.config.MercadopagoAccessToken)
		if err != nil {
			return nil, fmt.Errorf("mp config error: %w", err)
		}
		client := preferenceMp.NewClient(cfg)
		s.mercadopagoClient = NewDefaultMercadopagoClient(client)
	}

	var items []preferenceMp.ItemRequest
	for _, item := range order.Items {
		items = append(items, preferenceMp.ItemRequest{
			Title:       item.Product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   item.Price,
			Description: item.Product.Description,
		})
	}

	request := preferenceMp.Request{
		Items:             items,
		ExternalReference: order.OrderNumber,
		Payer: &preferenceMp.PayerRequest{
			Email: order.User.Email,
			Name:  order.User.FirstName + " " + order.User.LastName,
		},
	}

	paymentCreated, err := s.mercadopagoClient.CreatePreference(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("mercadopago api error: %w", err)
	}

	return paymentCreated, nil
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
	currentPayment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("error finding payment: %w", err)
	}

	if !paymentStatus.IsValidTransition(currentPayment.Status, status) {
		return nil, fmt.Errorf("invalid payment status transition from %s to %s", currentPayment.Status, status)
	}

	currentPayment.Status = status

	// Si el pago fue aprobado, actualizar estado de orden
	if status == paymentStatus.Approved {
		now := time.Now()
		currentPayment.ApprovedAt = &now

		// Actualizar orden a CONFIRMED
		if err := s.orderRepo.UpdateStatus(currentPayment.OrderID, orderStatus.Confirmed); err != nil {
			return nil, fmt.Errorf("error updating order status: %w", err)
		}
	}

	if err := s.paymentRepo.Update(currentPayment); err != nil {
		return nil, fmt.Errorf("error updating payment: %w", err)
	}

	return s.toPaymentResponse(currentPayment), nil
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
