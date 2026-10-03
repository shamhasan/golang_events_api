package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name        string    `json : "name" binding:"required"`
	Description string    `json : "description" binding:"required"`
	Location    string    `json : "location" binding:"required"`
	Image       string    `json : "image" binding:"required"`
	ImageId     string    `json : "imageId" binding:"required"`
	UserId      int       `json : "userId"`
	User        User      `gorm:"foreignKey:UserId" json:"user"`
	DateTime    time.Time `json :"dateTime" binding:"required"`
}
