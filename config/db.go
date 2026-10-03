package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"sham.com/app-server/models"

	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URI")

	if dsn == "" {
		log.Fatal("DATABASE_URI environment variable is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal terkoneksi ke DB", err)
	}

	err = database.AutoMigrate(&models.Event{}, models.User{})

	if err != nil {
		log.Fatal("Gagal melakukan migrasi database", err)
	}

	DB = database
	log.Println("Berhasil terkoneksi ke DB")
}
