package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model

	ID          uint    `gorm:"primarykey;autoIncrement:true;sequence:products_id_seq"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description"`
	Price       float64 `json:"price" gorm:"not null"`
	Stock       int     `json:"stock" gorm:"not null"`
	IsActive    bool    `json:"is_active" gorm:"default:true"`

	// we can add more fields like images, categories, etc.
	// but for now we will keep it simple
}
