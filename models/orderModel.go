package models

import (
	"gorm.io/gorm"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	gorm.Model

	ID          uint        `gorm:"primarykey;autoIncrement:true;sequence:orders_id_seq" json:"id"`
	UserID      uint        `json:"user_id" gorm:"not null"`
	User        User        `json:"user"`
	Status      OrderStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint    `json:"order_id" gorm:"not null"`
	ProductID uint    `json:"product_id" gorm:"not null"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity" gorm:"not null"`
	Price     float64 `json:"price" gorm:"not null"` // Price at time of order
}
