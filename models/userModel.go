package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`

	// we can add more fields here like first name, last name, phone number, etc
	// but we will keep it simple for now
}
