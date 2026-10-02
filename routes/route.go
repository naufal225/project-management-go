package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/joho/godotenv"
	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/controllers"
	"github.com/naufal225/project-management-go/utils"
)

func Setup(
	app *fiber.App,
	uc *controllers.UserController,
	bc *controllers.BoardController,
	lc *controllers.ListController,
	cc *controllers.CardController) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app.Post("/v1/auth/register", uc.Register)
	app.Post("/v1/auth/login", uc.Login)

	//JWT protected routes
	api := app.Group("/api/v1", jwtware.New(jwtware.Config{
		SigningKey: []byte(config.AppConfig.JWTSecret),
		ContextKey: "user",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.Unauthorized(c, "Error Unauthorized", err.Error())
		},
	}))

	userGroup := api.Group("/users")
	userGroup.Get("/page", uc.GetUserPagination)
	userGroup.Get("/:id", uc.GetUser)
	userGroup.Put("/:id", uc.UpdateUser)
	userGroup.Delete("/:id", uc.DeleteUser)

	boardGroup := api.Group("/boards")
	boardGroup.Post("/", bc.CreateBoard)
	boardGroup.Put("/:id", bc.UpdateBoard)
	boardGroup.Post("/:id/members", bc.AddBoardMembers)
	boardGroup.Delete("/:id/members", bc.RemoveBoardMembers)
	boardGroup.Get("/my", bc.GetMyBoardPaginate)
	boardGroup.Get("/:board_id/lists", lc.GetListOnBoard)
	boardGroup.Put("/:board_id/position", lc.UpdateListPosition);

	listGroup := api.Group("/lists")
	listGroup.Post("/", lc.CreateList)
	listGroup.Put("/:id", lc.UpdateList)
	listGroup.Delete("/:id", lc.DeleteList);

	cardGroup := api.Group("/cards")
	cardGroup.Post("/", cc.CreateCard)
	cardGroup.Put("/:id", cc.UpdateCard)
	cardGroup.Delete("/:id", cc.DeleteCard)
	cardGroup.Get("/:id", cc.GetCardDetail)

	cardGroup.Post("/:id/labels", cc.AddCardLabel)
	cardGroup.Delete("/:id/labels", cc.RemoveCardLabel)

	cardGroup.Post("/:id/attachments", cc.UploadAttachment)
	cardGroup.Get("/:id/attachments", cc.GetAttachment)
	cardGroup.Delete("/:card_id/attachments/:attachment_id", cc.DeleteAttachment)

	cardGroup.Post("/:id/assignees", cc.AddCardAssignee)
	cardGroup.Delete("/:id/assignees", cc.RemoveCardAssignee)
}
