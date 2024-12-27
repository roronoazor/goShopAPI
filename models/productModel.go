package models

import (
	"time"
)

type Product struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`

	// we can add more fields like images, categories, etc.
	// but for now we will keep it simple
}
