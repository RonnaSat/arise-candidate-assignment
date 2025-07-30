package models

import (
	"time"
)

type Order struct {
	ID            uint        `json:"id" gorm:"primaryKey"`
	TransactionID string      `json:"transaction_id" gorm:"unique;not null"`
	OrderItems    []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
	TotalAmount   float64     `json:"total_amount" gorm:"not null"`
	Status        string      `json:"status" gorm:"default:'pending'"` // pending, confirmed, shipped, delivered, cancelled
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OrderID   uint      `json:"order_id" gorm:"not null"`
	ProductID uint      `json:"product_id" gorm:"not null"`
	Product   Product   `json:"product" gorm:"foreignKey:ProductID"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
