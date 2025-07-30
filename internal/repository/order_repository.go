package repository

import (
	"gorepositorytest/internal/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	GetAll() ([]models.Order, error)
	GetByTransactionID(transactionID string) (*models.Order, error)
	Create(order *models.Order) error
	Update(order *models.Order) error
	UpdateStatus(orderID uint, status string) error
}

type postgresOrderRepository struct {
	db *gorm.DB
}

func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) GetAll() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Find(&orders).Error
	return orders, err
}

func (r *postgresOrderRepository) GetByTransactionID(transactionID string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("transaction_id = ?", transactionID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *postgresOrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *postgresOrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *postgresOrderRepository) UpdateStatus(orderID uint, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}
