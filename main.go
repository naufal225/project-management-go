package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/controllers"
	"github.com/naufal225/project-management-go/database/seed"
	"github.com/naufal225/project-management-go/repositories"
	"github.com/naufal225/project-management-go/routes"
	"github.com/naufal225/project-management-go/services"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	seed.SeedAdmin()

	app := fiber.New()

	userRepo := repositories.NewUserRepository()
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	boardMemberRepo := repositories.NewBoardMemberRepository()

	boardRepo := repositories.NewBoardRepository()
	boardService := services.NewBoardService(boardRepo, userRepo, boardMemberRepo)
	boardController := controllers.NewBoardController(boardService)

	listPositionRepo := repositories.NewListPositionRepository()
	listRepo := repositories.NewListRepository()
	listService := services.NewListService(listRepo, boardRepo, listPositionRepo)
	listController := controllers.NewListController(listService)

	cardRepo := repositories.NewCardRepository()

	attachmentRepo := repositories.NewAttachmentRepository(&gorm.DB{})
	attachmentService := services.NewAttachmentService(attachmentRepo, cardRepo, userRepo)

	labelRepo := repositories.NewLabelRepository()

	cardService := services.NewCardService(cardRepo, listRepo, userRepo, labelRepo)
	cardController := controllers.NewCardController(cardService, attachmentService, userService, "")

	routes.Setup(app, userController, boardController, listController, cardController)

	port := config.AppConfig.AppPort
	log.Println("Server is running on port :", port)
	log.Fatal(app.Listen(":" + port))
}
