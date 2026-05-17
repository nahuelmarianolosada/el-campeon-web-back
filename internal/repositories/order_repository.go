package repositories

import (
	"log"

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
	log.Printf("[orderRepository.Create] INFO: Creating order - orderNumber=%s, userID=%d, total=%.2f", order.OrderNumber, order.UserID, order.Total)

	if err := r.db.Create(order).Error; err != nil {
		log.Printf("[orderRepository.Create] ERROR: Failed to create order - orderNumber=%s: %v", order.OrderNumber, err)
		return err
	}
	log.Printf("[orderRepository.Create] INFO: Order created successfully - orderID=%d, orderNumber=%s", order.ID, order.OrderNumber)
	return nil
}

func (r *orderRepository) FindByID(id uint) (*models.Order, error) {
	log.Printf("[orderRepository.FindByID] INFO: Retrieving order - orderID=%d", id)

	var order models.Order
	if err := r.db.Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Images", orderImages).
		Preload("User").
		First(&order, id).Error; err != nil {
		log.Printf("[orderRepository.FindByID] ERROR: Failed to find order - orderID=%d: %v", id, err)
		return nil, err
	}
	log.Printf("[orderRepository.FindByID] INFO: Order found - orderID=%d, orderNumber=%s, status=%s, itemCount=%d", order.ID, order.OrderNumber, order.Status, len(order.Items))
	return &order, nil
}

func (r *orderRepository) FindByOrderNumber(orderNumber string) (*models.Order, error) {
	log.Printf("[orderRepository.FindByOrderNumber] INFO: Retrieving order by number - orderNumber=%s", orderNumber)

	var order models.Order
	if err := r.db.Preload("Items").Preload("Items.Product").Preload("Items.Product.Images", orderImages).Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		log.Printf("[orderRepository.FindByOrderNumber] ERROR: Failed to find order - orderNumber=%s: %v", orderNumber, err)
		return nil, err
	}
	log.Printf("[orderRepository.FindByOrderNumber] INFO: Order found - orderID=%d, orderNumber=%s, status=%s", order.ID, orderNumber, order.Status)
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint, limit, offset int) ([]models.Order, error) {
	log.Printf("[orderRepository.FindByUserID] INFO: Retrieving orders by user - userID=%d, limit=%d, offset=%d", userID, limit, offset)

	var orders []models.Order
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Preload("Items.Product").Preload("Items.Product.Images", orderImages).Find(&orders).Error; err != nil {
		log.Printf("[orderRepository.FindByUserID] ERROR: Failed to retrieve orders - userID=%d: %v", userID, err)
		return nil, err
	}
	log.Printf("[orderRepository.FindByUserID] INFO: Orders retrieved - userID=%d, orderCount=%d", userID, len(orders))
	return orders, nil
}

func (r *orderRepository) Update(order *models.Order) error {
	log.Printf("[orderRepository.Update] INFO: Updating order - orderID=%d, status=%s, total=%.2f", order.ID, order.Status, order.Total)

	if err := r.db.Save(order).Error; err != nil {
		log.Printf("[orderRepository.Update] ERROR: Failed to update order - orderID=%d: %v", order.ID, err)
		return err
	}
	log.Printf("[orderRepository.Update] INFO: Order updated successfully - orderID=%d", order.ID)
	return nil
}

func (r *orderRepository) Delete(id uint) error {
	log.Printf("[orderRepository.Delete] INFO: Deleting order - orderID=%d", id)

	if err := r.db.Delete(&models.Order{}, id).Error; err != nil {
		log.Printf("[orderRepository.Delete] ERROR: Failed to delete order - orderID=%d: %v", id, err)
		return err
	}
	log.Printf("[orderRepository.Delete] INFO: Order deleted successfully - orderID=%d", id)
	return nil
}

func (r *orderRepository) AddItem(orderID uint, item *models.OrderItem) error {
	log.Printf("[orderRepository.AddItem] INFO: Adding item to order - orderID=%d, productID=%d, quantity=%d", orderID, item.ProductID, item.Quantity)

	item.OrderID = orderID
	if err := r.db.Create(item).Error; err != nil {
		log.Printf("[orderRepository.AddItem] ERROR: Failed to add item - orderID=%d, productID=%d: %v", orderID, item.ProductID, err)
		return err
	}
	log.Printf("[orderRepository.AddItem] INFO: Item added successfully - orderID=%d, itemID=%d", orderID, item.ID)
	return nil
}

func (r *orderRepository) FindAll(limit, offset int) ([]models.Order, error) {
	log.Printf("[orderRepository.FindAll] INFO: Listing all orders - limit=%d, offset=%d", limit, offset)

	var orders []models.Order
	if err := r.db.Limit(limit).Offset(offset).Preload("Items.Product").Preload("Items.Product.Images", orderImages).Find(&orders).Error; err != nil {
		log.Printf("[orderRepository.FindAll] ERROR: Failed to list orders: %v", err)
		return nil, err
	}
	log.Printf("[orderRepository.FindAll] INFO: Orders listed - orderCount=%d", len(orders))
	return orders, nil
}

func (r *orderRepository) UpdateStatus(orderID uint, status string) error {
	log.Printf("[orderRepository.UpdateStatus] INFO: Updating order status - orderID=%d, newStatus=%s", orderID, status)

	if err := r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error; err != nil {
		log.Printf("[orderRepository.UpdateStatus] ERROR: Failed to update status - orderID=%d: %v", orderID, err)
		return err
	}
	log.Printf("[orderRepository.UpdateStatus] INFO: Order status updated - orderID=%d, status=%s", orderID, status)
	return nil
}
