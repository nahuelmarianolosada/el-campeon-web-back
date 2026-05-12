package repositories

import (
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
	return r.db.Create(payment).Error
}

func (r *paymentRepository) FindByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByOrderID(orderID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByMercadopagoPaymentID(mercadopagoPaymentID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.Where("mercadopago_payment_id = ?", mercadopagoPaymentID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByOrderNumber(orderNumber string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.Joins("Order").Where("`Order`.order_number = ?", orderNumber).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) FindByUserID(userID uint, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) ListAll(limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.db.Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) UpdateStatus(paymentID uint, status string) error {
	return r.db.Model(&models.Payment{}).Where("id = ?", paymentID).Update("status", status).Error
}
