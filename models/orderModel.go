package models

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	}
	return false
}

// Controls changing of order status
func (s OrderStatus) ValidateTransition(newStatus OrderStatus) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: must be one of [pending, processing, shipped, delivered, cancelled]")
	}

	// we can control changing of order status here
	// For example if an order had been completed and payment made, we can't cancel it
	// etc

	// we can more rules as needed in this function
	if s == StatusCancelled {
		return fmt.Errorf("cannot change status of cancelled order")
	}

	// Can't change status of delivered orders
	if s == StatusDelivered {
		return fmt.Errorf("cannot change status of delivered order")
	}

	return nil
}

type Order struct {
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" gorm:"index"`
	ID          uint        `gorm:"primarykey;autoIncrement:true;sequence:orders_id_seq" json:"id"`
	UserID      uint        `json:"user_id" gorm:"not null"`
	User        User        `json:"user"`
	Status      OrderStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	ID        uint       `gorm:"primarykey;autoIncrement:true;sequence:order_items_id_seq" json:"id"`
	OrderID   uint       `json:"order_id" gorm:"not null"`
	ProductID uint       `json:"product_id" gorm:"not null"`
	Product   Product    `json:"product"`
	Quantity  int        `json:"quantity" gorm:"not null"`
	Price     float64    `json:"price" gorm:"not null"` // price at time of order
}
