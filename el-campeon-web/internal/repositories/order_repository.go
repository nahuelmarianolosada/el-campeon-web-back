package repositories

import (
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *models.Order) error
	FindByID(id uint) (*models.Order, error)
	FindByOrderNumber(orderNumber string) (*models.Order, error)
	FindByUserID(userID uint, limit, offset int) ([]models.Order, error)
	Update(order *models.Order) error
	Delete(id uint) error
	AddItem(orderID uint, item *models.OrderItem) error
	FindAll(limit, offset int) ([]models.Order, error)
	UpdateStatus(orderID uint, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Items").Preload("Items.Product").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Items").Preload("Items.Product").Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&models.Order{}, id).Error
}

func (r *orderRepository) AddItem(orderID uint, item *models.OrderItem) error {
	item.OrderID = orderID
	return r.db.Create(item).Error
}

func (r *orderRepository) FindAll(limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	if err := r.db.Limit(limit).Offset(offset).Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) UpdateStatus(orderID uint, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

