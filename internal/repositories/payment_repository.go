package repositories

import (
	"log"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	FindByID(id uint) (*models.Payment, error)
	FindByTransactionID(transactionID string) (*models.Payment, error)
	FindByOrderID(orderID uint) (*models.Payment, error)
	FindByMercadopagoPaymentID(mercadopagoPaymentID string) (*models.Payment, error)
	FindByOrderNumber(orderNumber string) (*models.Payment, error)
	Update(payment *models.Payment) error
	FindByUserID(userID uint, limit, offset int) ([]models.Payment, error)
	ListAll(limit, offset int) ([]models.Payment, error)
	UpdateStatus(paymentID uint, status string) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	log.Printf("[paymentRepository.Create] INFO: Creating payment - orderID=%d, amount=%.2f, method=%s", payment.OrderID, payment.Amount, payment.PaymentMethod)
	if err := r.db.Create(payment).Error; err != nil {
		log.Printf("[paymentRepository.Create] ERROR: Failed to create payment - orderID=%d: %v", payment.OrderID, err)
		return err
	}
	log.Printf("[paymentRepository.Create] INFO: Payment created successfully - paymentID=%d, transactionID=%s", payment.ID, payment.TransactionID)
	return nil
}

func (r *paymentRepository) FindByID(id uint) (*models.Payment, error) {
	log.Printf("[paymentRepository.FindByID] INFO: Retrieving payment - paymentID=%d", id)
	var payment models.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		log.Printf("[paymentRepository.FindByID] ERROR: Failed to find payment - paymentID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByID] INFO: Payment found - paymentID=%d, status=%s", id, payment.Status)
	return &payment, nil
}

func (r *paymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	log.Printf("[paymentRepository.FindByTransactionID] INFO: Retrieving payment - transactionID=%s", transactionID)
	var payment models.Payment
	if err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		log.Printf("[paymentRepository.FindByTransactionID] ERROR: Failed to find payment - transactionID=%s: %v", transactionID, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByTransactionID] INFO: Payment found - paymentID=%d, transactionID=%s", payment.ID, transactionID)
	return &payment, nil
}

func (r *paymentRepository) FindByOrderID(orderID uint) (*models.Payment, error) {
	log.Printf("[paymentRepository.FindByOrderID] INFO: Retrieving payment - orderID=%d", orderID)
	var payment models.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		log.Printf("[paymentRepository.FindByOrderID] ERROR: Failed to find payment - orderID=%d: %v", orderID, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByOrderID] INFO: Payment found - paymentID=%d, orderID=%d", payment.ID, orderID)
	return &payment, nil
}

func (r *paymentRepository) FindByMercadopagoPaymentID(mercadopagoPaymentID string) (*models.Payment, error) {
	log.Printf("[paymentRepository.FindByMercadopagoPaymentID] INFO: Retrieving payment - mpPaymentID=%s", mercadopagoPaymentID)
	var payment models.Payment
	if err := r.db.Where("mercadopago_payment_id = ?", mercadopagoPaymentID).First(&payment).Error; err != nil {
		log.Printf("[paymentRepository.FindByMercadopagoPaymentID] ERROR: Failed to find payment - mpPaymentID=%s: %v", mercadopagoPaymentID, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByMercadopagoPaymentID] INFO: Payment found - paymentID=%d, mpPaymentID=%s", payment.ID, mercadopagoPaymentID)
	return &payment, nil
}

func (r *paymentRepository) FindByOrderNumber(orderNumber string) (*models.Payment, error) {
	log.Printf("[paymentRepository.FindByOrderNumber] INFO: Retrieving payment - orderNumber=%s", orderNumber)
	var payment models.Payment
	if err := r.db.Joins("Order").Where("`Order`.order_number = ?", orderNumber).First(&payment).Error; err != nil {
		log.Printf("[paymentRepository.FindByOrderNumber] ERROR: Failed to find payment - orderNumber=%s: %v", orderNumber, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByOrderNumber] INFO: Payment found - paymentID=%d, orderNumber=%s", payment.ID, orderNumber)
	return &payment, nil
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	log.Printf("[paymentRepository.Update] INFO: Updating payment - paymentID=%d, status=%s", payment.ID, payment.Status)
	if err := r.db.Save(payment).Error; err != nil {
		log.Printf("[paymentRepository.Update] ERROR: Failed to update payment - paymentID=%d: %v", payment.ID, err)
		return err
	}
	log.Printf("[paymentRepository.Update] INFO: Payment updated successfully - paymentID=%d", payment.ID)
	return nil
}

func (r *paymentRepository) FindByUserID(userID uint, limit, offset int) ([]models.Payment, error) {
	log.Printf("[paymentRepository.FindByUserID] INFO: Retrieving payments - userID=%d, limit=%d, offset=%d", userID, limit, offset)
	var payments []models.Payment
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		log.Printf("[paymentRepository.FindByUserID] ERROR: Failed to retrieve payments - userID=%d: %v", userID, err)
		return nil, err
	}
	log.Printf("[paymentRepository.FindByUserID] INFO: Payments retrieved - userID=%d, count=%d", userID, len(payments))
	return payments, nil
}

func (r *paymentRepository) ListAll(limit, offset int) ([]models.Payment, error) {
	log.Printf("[paymentRepository.ListAll] INFO: Listing all payments - limit=%d, offset=%d", limit, offset)
	var payments []models.Payment
	if err := r.db.Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		log.Printf("[paymentRepository.ListAll] ERROR: Failed to list payments: %v", err)
		return nil, err
	}
	log.Printf("[paymentRepository.ListAll] INFO: Payments listed - count=%d", len(payments))
	return payments, nil
}

func (r *paymentRepository) UpdateStatus(paymentID uint, status string) error {
	log.Printf("[paymentRepository.UpdateStatus] INFO: Updating payment status - paymentID=%d, status=%s", paymentID, status)
	if err := r.db.Model(&models.Payment{}).Where("id = ?", paymentID).Update("status", status).Error; err != nil {
		log.Printf("[paymentRepository.UpdateStatus] ERROR: Failed to update payment status - paymentID=%d: %v", paymentID, err)
		return err
	}
	log.Printf("[paymentRepository.UpdateStatus] INFO: Payment status updated successfully - paymentID=%d, status=%s", paymentID, status)
	return nil
}
