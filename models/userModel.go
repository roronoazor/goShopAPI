package models

import "gorm.io/gorm"

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleCustomer UserRole = "customer"
)

type User struct {
	gorm.Model
	Username string   `json:"username" gorm:"unique"`
	Email    string   `json:"email" gorm:"unique"`
	Password string   `json:"password"`
	Role     UserRole `json:"role" gorm:"type:varchar(20);default:'customer'"`
	// we can add more fields here like first name, last name, phone number, etc
	// but we will keep it simple for now
}
