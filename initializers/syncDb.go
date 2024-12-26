package initializers

import (
	"log"

	"github.com/roronoazor/goShopAPI/models"
)

func SyncDb() {
	// Add all your models here
	err := DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)

	if err != nil {
		log.Fatal("Failed to sync database:", err)
	}

	log.Println("Database synced successfully")
}
