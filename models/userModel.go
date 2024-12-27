package models

import (
	"fmt"

	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleCustomer UserRole = "customer"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleCustomer, "":
		return true
	}
	return false
}

// ValidateRole checks if the role is valid and allowed for signup
func (r UserRole) ValidateForSignup() error {
	if !r.IsValid() {
		return fmt.Errorf("invalid role: must be either 'customer' or 'admin'")
	}
	if r == UserRoleAdmin {
		return fmt.Errorf("admin role cannot be set during signup")
	}
	return nil
}

type User struct {
	gorm.Model
	ID       uint     `gorm:"primarykey;autoIncrement:true;sequence:users_id_seq" json:"id"`
	Username string   `json:"username" gorm:"unique"`
	Email    string   `json:"email" gorm:"unique"`
	Password string   `json:"password"`
	Role     UserRole `json:"role" gorm:"type:varchar(20);default:'customer'"`
	// we can add more fields here like first name, last name, phone number, etc
	// but we will keep it simple for now
}
