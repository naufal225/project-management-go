package seed

import (
	"log"

	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/utils"
)

func SeedAdmin() {
	password, _ := utils.HashPassword("password")

	admin := models.User{
		Name:     "Super admin",
		Email:    "admin@example.com",
		Password: password,
		Role:     "admin",
	}

	if err := config.DB.FirstOrCreate(&admin, models.User{Email: admin.Email, Name: admin.Name}).Error; err != nil {
		log.Println("Failed to seed admin.", err)
	} else {
		log.Println("admin user seeded.")
	}
	
}
