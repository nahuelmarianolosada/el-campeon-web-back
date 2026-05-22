package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	CreateGuestPayment(ctx context.Context, req *models.CreateGuestPaymentRequest) (*models.PaymentResponse, error)
	GetPaymentByID(id uint) (*models.PaymentResponse, error)
	GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error)
	GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error)
	UpdatePaymentStatus(paymentID uint, status string) (*models.PaymentResponse, error)
	ProcessMercadopagoWebhook(ctx context.Context, webhook *models.MercadopagoWebhookRequest, xSignature string) error
	ListAllPayments(limit, offset int) ([]models.PaymentResponse, error)
}

type paymentService struct {
	paymentRepo       repositories.PaymentRepository
	orderRepo         repositories.OrderRepository
	config            *internalConfig.Config
	mercadopagoClient MercadopagoClient
	webhookValidator  *WebhookValidator
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
		paymentRepo:       paymentRepo,
		orderRepo:         orderRepo,
		config:            cfg,
		mercadopagoClient: nil, // Will be set via SetMercadopagoClient or use default
		webhookValidator:  validator,
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
		webhookValidator:  validator,
	}
}

func (s *paymentService) SetMercadopagoClient(client MercadopagoClient) {
	s.mercadopagoClient = client
}

func (s *paymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	log.Printf("[paymentService.CreatePayment] INFO: Starting payment creation - orderID=%d, amount=%.2f, paymentMethod=%s", req.OrderID, req.Amount, req.PaymentMethod)

	// Obtener la orden
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to find order - orderID=%d: %v", req.OrderID, err)
		return nil, fmt.Errorf("error finding order: %w", err)
	}
	log.Printf("[paymentService.CreatePayment] INFO: Order found - orderID=%d, status=%s, total=%.2f", order.ID, order.Status, order.Total)

	// Validar que la orden no esté cancelada
	if order.Status == orderStatus.Cancelled {
		log.Printf("[paymentService.CreatePayment] WARNING: Attemptto create payment for cancelled order - orderID=%d", req.OrderID)
		return nil, fmt.Errorf("cannot create payment for cancelled order")
	}

	// Verificar que el monto coincide
	if req.Amount != order.Total {
		log.Printf("[paymentService.CreatePayment] ERROR: Amount mismatch - orderID=%d, expected=%.2f, received=%.2f", req.OrderID, order.Total, req.Amount)
		return nil, fmt.Errorf("payment amount does not match order total. expected: %.2f, got: %.2f", order.Total, req.Amount)
	}

	// Crear pago
	transactionID := s.generateTransactionID()
	var userID *uint
	if order.UserID != nil {
		userID = order.UserID
	}
	payment := &models.Payment{
		TransactionID: transactionID,
		OrderID:       req.OrderID,
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      "ARS",
		Status:        paymentStatus.Pending,
		PaymentMethod: req.PaymentMethod,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to create payment - orderID=%d: %v", req.OrderID, err)
		return nil, fmt.Errorf("error creating payment: %w", err)
	}
	log.Printf("[paymentService.CreatePayment] INFO: Payment created - paymentID=%d, transactionID=%s, status=%s", payment.ID, transactionID, payment.Status)

	// Para pagos en efectivo, marcar como pendiente de confirmación
	if req.PaymentMethod == "CASH" {
		payment.Status = paymentStatus.Pending
		// No procesamos con MercadoPago para pagos en efectivo
		if err := s.paymentRepo.Update(payment); err != nil {
			log.Printf("[paymentService.CreatePayment] ERROR: Failed to update cash payment - paymentID=%d: %v", payment.ID, err)
			return nil, fmt.Errorf("error updating payment: %w", err)
		}
		log.Printf("[paymentService.CreatePayment] INFO: Cash payment confirmed - paymentID=%d, amount=%.2f", payment.ID, req.Amount)
		return s.toPaymentResponse(payment), nil
	}

	// Para pagos con MercadoPago, crear preference
	log.Printf("[paymentService.CreatePayment] INFO: Processing MercadoPago payment - paymentID=%d", payment.ID)
	executedPayment, err := s.ExecutePayment(ctx, *order, req.PaymentMethod)
	if err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to execute MercadoPago payment - orderID=%d: %v", req.OrderID, err)
		return nil, fmt.Errorf("error executing payment: %w", err)
	}

	executedPaymentByte, err := json.Marshal(executedPayment)
	if err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to marshal executed payment - paymentID=%d: %v", payment.ID, err)
		return nil, fmt.Errorf("error marshaling executed payment: %w", err)
	}

	payment.MercadopagoPreferenceID = executedPayment.ID
	payment.MercadopagoData = datatypes.JSONMap{
		"preference": string(executedPaymentByte),
	}
	payment.Status = paymentStatus.Pending // En Mercado Pago, el pago sigue siendo PENDING hasta confirmación de webhook

	if err := s.paymentRepo.Update(payment); err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to update payment with MP data - paymentID=%d: %v", payment.ID, err)
		return nil, fmt.Errorf("error updating payment with MP data: %w", err)
	}

	// Actualizar orden a CONFIRMED
	order.Status = orderStatus.Confirmed
	order.UpdatedAt = time.Now()
	if err := s.orderRepo.Update(order); err != nil {
		log.Printf("[paymentService.CreatePayment] ERROR: Failed to update order status - orderID=%d: %v", order.ID, err)
		return nil, fmt.Errorf("error updating order status: %w", err)
	}
	log.Printf("[paymentService.CreatePayment] INFO: MercadoPago payment prepared - paymentID=%d, preferenceID=%s, amount=%.2f", payment.ID, executedPayment.ID, req.Amount)

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
		Items:               items,
		ExternalReference:   order.OrderNumber,
		Payer:               &payer,
		StatementDescriptor: fmt.Sprintf("Campeon %s", order.OrderNumber),
	}

	paymentCreated, err := s.mercadopagoClient.CreatePreference(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("mercadopago api error: %w", err)
	}

	return paymentCreated, nil
}

// CreateGuestPayment crea un pago para una orden guest (sin usuario autenticado)
func (s *paymentService) CreateGuestPayment(ctx context.Context, req *models.CreateGuestPaymentRequest) (*models.PaymentResponse, error) {
	log.Printf("[paymentService.CreateGuestPayment] INFO: Starting guest payment creation - orderID=%d, email=%s, amount=%.2f", req.OrderID, req.Email, req.Amount)

	// Obtener la orden
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to find order - orderID=%d: %v", req.OrderID, err)
		return nil, fmt.Errorf("error finding order: %w", err)
	}

	// Validar que es orden guest y el email coincide
	if order.GuestEmail == "" {
		log.Printf("[paymentService.CreateGuestPayment] WARNING: Order is not a guest order - orderID=%d", req.OrderID)
		return nil, fmt.Errorf("this order is not a guest order")
	}

	if order.GuestEmail != req.Email {
		log.Printf("[paymentService.CreateGuestPayment] WARNING: Email mismatch - orderID=%d, expected=%s, got=%s", req.OrderID, order.GuestEmail, req.Email)
		return nil, fmt.Errorf("email does not match order")
	}

	// Validar que la orden no esté cancelada
	if order.Status == orderStatus.Cancelled {
		log.Printf("[paymentService.CreateGuestPayment] WARNING: Attempt to create payment for cancelled order - orderID=%d", req.OrderID)
		return nil, fmt.Errorf("cannot create payment for cancelled order")
	}

	// Verificar que el monto coincide
	if req.Amount != order.Total {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Amount mismatch - orderID=%d, expected=%.2f, received=%.2f", req.OrderID, order.Total, req.Amount)
		return nil, fmt.Errorf("payment amount does not match order total")
	}

	// Crear pago (UserID será nil para guest)
	transactionID := s.generateTransactionID()
	payment := &models.Payment{
		TransactionID: transactionID,
		OrderID:       req.OrderID,
		UserID:        order.UserID, // Maintain consistency with order's UserID (which is nil for guests)
		Amount:        req.Amount,
		Currency:      "ARS",
		Status:        paymentStatus.Pending,
		PaymentMethod: req.PaymentMethod,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to create payment: %v", err)
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	// Para pagos en efectivo
	if req.PaymentMethod == "CASH" {
		log.Printf("[paymentService.CreateGuestPayment] INFO: Cash payment created - paymentID=%d, amount=%.2f", payment.ID, req.Amount)
		return s.toPaymentResponse(payment), nil
	}

	// Para pagos con MercadoPago
	log.Printf("[paymentService.CreateGuestPayment] INFO: Processing MercadoPago payment - paymentID=%d", payment.ID)
	executedPayment, err := s.ExecuteGuestPayment(ctx, *order, req.PaymentMethod, req.Email)
	if err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to execute MercadoPago payment: %v", err)
		return nil, fmt.Errorf("error executing payment: %w", err)
	}

	executedPaymentByte, err := json.Marshal(executedPayment)
	if err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to marshal executed payment: %v", err)
		return nil, fmt.Errorf("error marshaling executed payment: %w", err)
	}

	payment.MercadopagoPreferenceID = executedPayment.ID
	payment.MercadopagoData = datatypes.JSONMap{
		"preference": string(executedPaymentByte),
	}
	payment.Status = paymentStatus.Pending

	if err := s.paymentRepo.Update(payment); err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to update payment: %v", err)
		return nil, fmt.Errorf("error updating payment: %w", err)
	}

	// Actualizar orden a CONFIRMED
	order.Status = orderStatus.Confirmed
	order.UpdatedAt = time.Now()
	if err := s.orderRepo.Update(order); err != nil {
		log.Printf("[paymentService.CreateGuestPayment] ERROR: Failed to update order status: %v", err)
		return nil, fmt.Errorf("error updating order status: %w", err)
	}

	log.Printf("[paymentService.CreateGuestPayment] INFO: Guest payment prepared - paymentID=%d, preferenceID=%s", payment.ID, executedPayment.ID)
	return s.toPaymentResponse(payment), nil
}

// ExecuteGuestPayment es similar a ExecutePayment pero usa email de guest en lugar de usuario
func (s *paymentService) ExecuteGuestPayment(ctx context.Context, order models.Order, paymentMethod, guestEmail string) (*preferenceMp.Response, error) {
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

	// Usar email de guest
	var payer preferenceMp.PayerRequest
	payer.Email = guestEmail

	request := preferenceMp.Request{
		Items:               items,
		ExternalReference:   order.OrderNumber,
		Payer:               &payer,
		StatementDescriptor: fmt.Sprintf("Campeon %s", order.OrderNumber),
	}

	paymentCreated, err := s.mercadopagoClient.CreatePreference(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("mercadopago api error: %w", err)
	}

	return paymentCreated, nil
}

func (s *paymentService) GetPaymentByID(id uint) (*models.PaymentResponse, error) {
	log.Printf("[paymentService.GetPaymentByID] INFO: Retrieving payment - paymentID=%d", id)

	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		log.Printf("[paymentService.GetPaymentByID] ERROR: Failed to find payment - paymentID=%d: %v", id, err)
		return nil, fmt.Errorf("error finding payment: %w", err)
	}
	log.Printf("[paymentService.GetPaymentByID] INFO: Payment found - paymentID=%d, status=%s, amount=%.2f", payment.ID, payment.Status, payment.Amount)

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) GetPaymentsByUserID(userID uint, limit, offset int) ([]models.PaymentResponse, error) {
	log.Printf("[paymentService.GetPaymentsByUserID] INFO: Retrieving payments for user - userID=%d, limit=%d, offset=%d", userID, limit, offset)

	payments, err := s.paymentRepo.FindByUserID(userID, limit, offset)
	if err != nil {
		log.Printf("[paymentService.GetPaymentsByUserID] ERROR: Failed to find payments - userID=%d: %v", userID, err)
		return nil, fmt.Errorf("error finding payments: %w", err)
	}
	log.Printf("[paymentService.GetPaymentsByUserID] INFO: Payments retrieved - userID=%d, paymentCount=%d", userID, len(payments))

	var responses []models.PaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.toPaymentResponse(&payment))
	}

	return responses, nil
}

func (s *paymentService) GetPaymentByOrderID(orderID uint) (*models.PaymentResponse, error) {
	log.Printf("[paymentService.GetPaymentByOrderID] INFO: Retrieving payment for order - orderID=%d", orderID)

	payment, err := s.paymentRepo.FindByOrderID(orderID)
	if err != nil {
		log.Printf("[paymentService.GetPaymentByOrderID] ERROR: Failed to find payment - orderID=%d: %v", orderID, err)
		return nil, fmt.Errorf("error finding payment for order: %w", err)
	}
	log.Printf("[paymentService.GetPaymentByOrderID] INFO: Payment found - orderID=%d, paymentID=%d, status=%s", orderID, payment.ID, payment.Status)

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) UpdatePaymentStatus(paymentID uint, status string) (*models.PaymentResponse, error) {
	log.Printf("[paymentService.UpdatePaymentStatus] INFO: Starting payment status update - paymentID=%d, newStatus=%s", paymentID, status)

	currentPayment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		log.Printf("[paymentService.UpdatePaymentStatus] ERROR: Failed to find payment - paymentID=%d: %v", paymentID, err)
		return nil, fmt.Errorf("error finding payment: %w", err)
	}

	if !paymentStatus.IsValidTransition(currentPayment.Status, status) {
		log.Printf("[paymentService.UpdatePaymentStatus] WARNING: Invalid status transition - paymentID=%d, currentStatus=%s, requestedStatus=%s", paymentID, currentPayment.Status, status)
		return nil, fmt.Errorf("invalid payment status transition from %s to %s", currentPayment.Status, status)
	}

	currentPayment.Status = status

	// Si el pago fue aprobado, actualizar estado de orden
	if status == paymentStatus.Approved {
		now := time.Now()
		currentPayment.ApprovedAt = &now
		log.Printf("[paymentService.UpdatePaymentStatus] INFO: Payment approved - paymentID=%d", paymentID)

		// Actualizar orden a CONFIRMED
		if err := s.orderRepo.UpdateStatus(currentPayment.OrderID, orderStatus.Confirmed); err != nil {
			log.Printf("[paymentService.UpdatePaymentStatus] ERROR: Failed to update order status - orderID=%d: %v", currentPayment.OrderID, err)
			return nil, fmt.Errorf("error updating order status: %w", err)
		}
	}

	if err := s.paymentRepo.Update(currentPayment); err != nil {
		log.Printf("[paymentService.UpdatePaymentStatus] ERROR: Failed to update payment status - paymentID=%d: %v", paymentID, err)
		return nil, fmt.Errorf("error updating payment: %w", err)
	}
	log.Printf("[paymentService.UpdatePaymentStatus] INFO: Payment status updated - paymentID=%d, oldStatus=%s, newStatus=%s", paymentID, currentPayment.Status, status)

	return s.toPaymentResponse(currentPayment), nil
}

// ProcessMercadopagoWebhook procesa webhooks de MercadoPago
// Realiza validación de firma, obtiene detalles del pago desde MP API
// y actualiza los estados de pago y orden
func (s *paymentService) ProcessMercadopagoWebhook(ctx context.Context, webhook *models.MercadopagoWebhookRequest, xSignature string) error {
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Processing webhook - webhookType=%s, webhookDataID=%s", webhook.Type, webhook.Data.ID)

	// 1. Validar que sea un webhook de pago
	if webhook.Type != "payment" {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Ignoring non-payment webhook - webhookType=%s", webhook.Type)
		return nil // Ignorar otros tipos de eventos
	}

	// 2. Validar la firma del webhook
	if !s.webhookValidator.ValidateSignature(xSignature, webhook.Type, webhook.Data.ID, s.config.MercadopagoAccessToken) {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Invalid webhook signature - webhookDataID=%s", webhook.Data.ID)
		return fmt.Errorf("invalid webhook signature")
	}
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Webhook signature validated - webhookDataID=%s", webhook.Data.ID)

	// 2.5. Inicializar cliente si no existe
	if s.mercadopagoClient == nil {
		cfg, err := config.New(s.config.MercadopagoAccessToken)
		if err != nil {
			log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to create MP config: %v", err)
			return fmt.Errorf("mp config error: %w", err)
		}
		client := preferenceMp.NewClient(cfg)
		s.mercadopagoClient = NewDefaultMercadopagoClient(client, s.config.MercadopagoAccessToken)
	}

	// 3. Obtener los detalles completos del pago desde MercadoPago API
	paymentDetails, err := s.mercadopagoClient.GetPaymentDetails(ctx, webhook.Data.ID)
	if err != nil {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to fetch payment details from MP - webhookDataID=%s: %v", webhook.Data.ID, err)
		return fmt.Errorf("error fetching payment details from mercadopago: %w", err)
	}
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Payment details retrieved from MP - paymentID=%d, mpStatus=%s, amount=%.2f", paymentDetails.ID, paymentDetails.Status, paymentDetails.TransactionAmount)

	// 4. Buscar el pago local usando el MercadopagoPaymentID
	payment, err := s.paymentRepo.FindByOrderNumber(paymentDetails.ExternalReference)
	if err != nil {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to find local payment - externalRef=%s: %v", paymentDetails.ExternalReference, err)
		return fmt.Errorf("error finding payment: %w", err)
	}
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Local payment found - paymentID=%d, orderNumber=%s", payment.ID, paymentDetails.ExternalReference)

	// 5. Verificar que los montos coincidan
	if paymentDetails.TransactionAmount != payment.Amount {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Amount mismatch - paymentID=%d, expected=%.2f, received=%.2f", payment.ID, payment.Amount, paymentDetails.TransactionAmount)
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
		log.Printf("[paymentService.ProcessMercadopagoWebhook] WARNING: Payment rejected - paymentID=%d, reason=%s", payment.ID, rejectedReason)
	}

	// Si fue aprobado, establecer la fecha de aprobación
	if paymentDetails.Status == "approved" {
		now := time.Now()
		payment.ApprovedAt = &now
		log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Payment approved - paymentID=%d", payment.ID)
	}

	// Guardar los datos completos de MercadoPago
	mpDataBytes, err := json.Marshal(paymentDetails)
	if err != nil {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to marshal payment details - paymentID=%d: %v", payment.ID, err)
		return fmt.Errorf("error marshaling payment details: %w", err)
	}

	payment.MercadopagoData = datatypes.JSONMap{
		"payment_details":  string(mpDataBytes),
		"webhook_received": time.Now().Format(time.RFC3339),
	}

	// 8. Actualizar el pago en la base de datos
	if err := s.paymentRepo.Update(payment); err != nil {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to update payment - paymentID=%d: %v", payment.ID, err)
		return fmt.Errorf("error updating payment: %w", err)
	}
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Payment updated in DB - paymentID=%d, newStatus=%s", payment.ID, newStatus)

	// 9. Actualizar el estado de la orden según el estado del pago
	order, err := s.orderRepo.FindByID(payment.OrderID)
	if err != nil {
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to find order - orderID=%d: %v", payment.OrderID, err)
		return fmt.Errorf("error finding order: %w", err)
	}

	oldOrderStatus := order.Status
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
		log.Printf("[paymentService.ProcessMercadopagoWebhook] ERROR: Failed to update order - orderID=%d: %v", order.ID, err)
		return fmt.Errorf("error updating order status: %w", err)
	}
	log.Printf("[paymentService.ProcessMercadopagoWebhook] INFO: Webhook processed successfully - paymentID=%d, newPaymentStatus=%s, oldOrderStatus=%s, newOrderStatus=%s", payment.ID, newStatus, oldOrderStatus, order.Status)

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
	log.Printf("[paymentService.ListAllPayments] INFO: Listing all payments - limit=%d, offset=%d", limit, offset)

	payments, err := s.paymentRepo.ListAll(limit, offset)
	if err != nil {
		log.Printf("[paymentService.ListAllPayments] ERROR: Failed to list payments: %v", err)
		return nil, fmt.Errorf("error listing payments: %w", err)
	}
	log.Printf("[paymentService.ListAllPayments] INFO: Payments listed - paymentCount=%d", len(payments))

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
