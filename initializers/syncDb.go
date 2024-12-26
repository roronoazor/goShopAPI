package initializers

import "github.com/roronoazor/goShopAPI/models"

func SyncDb() {
	DB.AutoMigrate(&models.User{})
}
