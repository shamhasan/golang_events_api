package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name" `
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" `
	Events   []Event
}
