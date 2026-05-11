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
	ProcessMercadopagoWebhook(ctx context.Context, webhook *models.MercadopagoWebhookRequest, xSignature string) error
	ListAllPayments(limit, offset int) ([]models.PaymentResponse, error)
}

type paymentService struct {
	paymentRepo         repositories.PaymentRepository
	orderRepo           repositories.OrderRepository
	config              *internalConfig.Config
	mercadopagoClient   MercadopagoClient
	webhookValidator    *WebhookValidator
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cfg *internalConfig.Config,
) PaymentService {
	validator := &WebhookValidator{}
	if cfg != nil {
		validator = NewWebhookValidator(cfg.MercadopagoPublicKey)
	}
	return &paymentService{
		paymentRepo:      paymentRepo,
		orderRepo:        orderRepo,
		config:           cfg,
		mercadopagoClient: nil, // Will be set via SetMercadopagoClient or use default
		webhookValidator: validator,
	}
}

func NewPaymentServiceWithClient(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cfg *internalConfig.Config,
	client MercadopagoClient,
) PaymentService {
	validator := &WebhookValidator{}
	if cfg != nil {
		validator = NewWebhookValidator(cfg.MercadopagoPublicKey)
	}
	return &paymentService{
		paymentRepo:       paymentRepo,
		orderRepo:         orderRepo,
		config:            cfg,
		mercadopagoClient: client,
		webhookValidator: validator,
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
		s.mercadopagoClient = NewDefaultMercadopagoClient(client, s.config.MercadopagoAccessToken)
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

	var payer preferenceMp.PayerRequest
	if order.User != nil {
		payer.Email = order.User.Email
		payer.Name = order.User.FirstName + " " + order.User.LastName
	}

	request := preferenceMp.Request{
		Items:             items,
		ExternalReference: order.OrderNumber,
		Payer: &payer,
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
// Realiza validación de firma, obtiene detalles del pago desde MP API
// y actualiza los estados de pago y orden
func (s *paymentService) ProcessMercadopagoWebhook(ctx context.Context, webhook *models.MercadopagoWebhookRequest, xSignature string) error {
	// 1. Validar que sea un webhook de pago
	if webhook.Type != "payment" {
		return nil // Ignorar otros tipos de eventos
	}

	// 2. Validar la firma del webhook
	if !s.webhookValidator.ValidateSignature(xSignature, webhook.Type, webhook.Data.ID, s.config.MercadopagoAccessToken) {
		return fmt.Errorf("invalid webhook signature")
	}

	// 3. Obtener los detalles completos del pago desde MercadoPago API
	paymentDetails, err := s.mercadopagoClient.GetPaymentDetails(ctx, webhook.Data.ID)
	if err != nil {
		return fmt.Errorf("error fetching payment details from mercadopago: %w", err)
	}

	// 4. Buscar el pago local usando el MercadopagoPaymentID
	payment, err := s.paymentRepo.FindByMercadopagoPaymentID(fmt.Sprintf("%d", paymentDetails.ID))
	if err != nil {
		return fmt.Errorf("error finding payment: %w", err)
	}

	// 5. Verificar que los montos coincidan
	if paymentDetails.TransactionAmount != payment.Amount {
		return fmt.Errorf("payment amount mismatch: expected %.2f, got %.2f", payment.Amount, paymentDetails.TransactionAmount)
	}

	// 6. Mapear el estado de MercadoPago a nuestro estado local
	newStatus := mapMercadopagoStatusToLocalStatus(paymentDetails.Status)
	rejectedReason := ""

	// 7. Actualizar el pago con los detalles de MercadoPago
	payment.MercadopagoPaymentID = fmt.Sprintf("%d", paymentDetails.ID)
	payment.Status = newStatus

	// Si fue rechazado, guardar el motivo
	if paymentDetails.Status == "rejected" {
		rejectedReason = paymentDetails.StatusDetail
		payment.RejectedReason = rejectedReason
	}

	// Si fue aprobado, establecer la fecha de aprobación
	if paymentDetails.Status == "approved" {
		now := time.Now()
		payment.ApprovedAt = &now
	}

	// Guardar los datos completos de MercadoPago
	mpDataBytes, err := json.Marshal(paymentDetails)
	if err != nil {
		return fmt.Errorf("error marshaling payment details: %w", err)
	}

	payment.MercadopagoData = datatypes.JSONMap{
		"payment_details": string(mpDataBytes),
		"webhook_received": time.Now().Format(time.RFC3339),
	}

	// 8. Actualizar el pago en la base de datos
	if err := s.paymentRepo.Update(payment); err != nil {
		return fmt.Errorf("error updating payment: %w", err)
	}

	// 9. Actualizar el estado de la orden según el estado del pago
	order, err := s.orderRepo.FindByID(payment.OrderID)
	if err != nil {
		return fmt.Errorf("error finding order: %w", err)
	}

	switch newStatus {
	case paymentStatus.Approved:
		// Cambiar orden a CONFIRMED
		order.Status = orderStatus.Confirmed
	case paymentStatus.Rejected:
		// Cambiar orden a CANCELLED
		order.Status = orderStatus.Cancelled
	case paymentStatus.Refunded:
		// Cambiar orden a CANCELLED
		order.Status = orderStatus.Cancelled
	}

	order.UpdatedAt = time.Now()
	if err := s.orderRepo.Update(order); err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}

	fmt.Printf("Successfully processed webhook for payment ID %s, status: %s\n", webhook.Data.ID, newStatus)
	return nil
}

// mapMercadopagoStatusToLocalStatus mapea los estados de MercadoPago a nuestros estados locales
func mapMercadopagoStatusToLocalStatus(mpStatus string) string {
	switch mpStatus {
	case "approved":
		return paymentStatus.Approved
	case "rejected":
		return paymentStatus.Rejected
	case "refunded":
		return paymentStatus.Refunded
	case "charged_back":
		return paymentStatus.Rejected // Tratar como rechazado
	case "pending":
		return paymentStatus.Pending
	case "cancelled":
		return paymentStatus.Cancelled
	case "in_process":
		return paymentStatus.Pending
	case "in_mediation":
		return paymentStatus.Pending
	default:
		return paymentStatus.Pending
	}
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
