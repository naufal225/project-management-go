package controllers

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	dtos "github.com/naufal225/project-management-go/DTOs"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/services"
	"github.com/naufal225/project-management-go/utils"
)

type BoardController struct {
	service services.BoardService
}

func NewBoardController(s services.BoardService) *BoardController {
	return &BoardController{service: s}
}

func (b *BoardController) CreateBoard(ctx *fiber.Ctx) error {
	var userID uuid.UUID
	var err error

	boardReq := new(dtos.CreateBoardRequest)
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	if err := ctx.BodyParser(&boardReq); err != nil {
		return utils.BadRequest(ctx, "Gagal membaca request", err.Error())
	}

	userID, err = uuid.Parse(claims["pub_id"].(string))
	if err != nil {
		return utils.BadRequest(ctx, "Public ID tidak valid", err.Error())
	}

	boardReq.OwnerPublicID = userID

	newBoard := models.Board{}
	if err := copier.Copy(&newBoard, &boardReq); err != nil {
		return utils.InternalServerError(ctx, "Gagal parsing data", err.Error())
	}

	if err := b.service.Create(&newBoard); err != nil {
		return utils.BadRequest(ctx, "Gagal menyimpan data", err.Error())
	}

	boardRes := new(dtos.BoardResponse)
	if err := copier.Copy(&boardRes, &newBoard); err != nil {
		return utils.InternalServerError(ctx, "Gagal parsing data", err.Error())
	}

	return utils.Success(ctx, "Berhasil membuat board", boardRes)
}

func (c *BoardController) UpdateBoard(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	board := new(models.Board)
	req := new(dtos.UpdateBoardRequest)

	if err := ctx.BodyParser(req); err != nil {
		return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
	}

	if _, err := uuid.Parse(publicID); err != nil {
		return utils.BadRequest(ctx, "ID tidak valid", err.Error())
	}

	existingBoard, err := c.service.GetByPublicID(publicID)

	if err != nil {
		return utils.NotFound(ctx, "Board tidak ditemukan", err.Error())
	}

	board.Title = req.Title
	board.Description = req.Description
	board.DueDate = req.DueDate

	board.InternalID = existingBoard.InternalID
	board.PublicID = existingBoard.PublicID
	board.OwnerID = existingBoard.OwnerID
	board.OwnerPublicID = existingBoard.OwnerPublicID
	board.CreatedAt = existingBoard.CreatedAt

	if err := c.service.Update(board); err != nil {
		return utils.BadRequest(ctx, "Gagal update board", err.Error())
	}

	return utils.Success(ctx, "Board berhasil diupdate", board)
}

func (c *BoardController) AddBoardMembers(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")

	var userIDs []string
	if err := ctx.BodyParser(&userIDs); err != nil {
		return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
	}

	if err := c.service.AddMembers(publicID, userIDs); err != nil {
		return utils.BadRequest(ctx, "Gagal menambahkan members", err.Error())
	}

	return utils.Success(ctx, "Berhasil menambahkan members", nil)
}

func (c *BoardController) RemoveBoardMembers(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")

	var userIDs []string
	if err := ctx.BodyParser(&userIDs); err != nil {
		return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
	}

	if err := c.service.RemoveMembers(publicID, userIDs); err != nil {
		return utils.BadRequest(ctx, "Gagal menghapus members", err.Error())
	}

	return utils.Success(ctx, "Berhasil menghapus members", nil)
}

func (c *BoardController) GetMyBoardPaginate(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["pub_id"].(string)

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("page", "10"))
	offset := (page -1) * limit

	filter := ctx.Query("filter", "")
	sort := ctx.Query("sort", "")

	boards, total, err := c.service.GetAllByUserPaginate(userID, filter, sort, limit, offset)
	
	if err != nil {
		return utils.InternalServerError(ctx, "Gagal mengambil data board", err.Error())
	}

	meta := utils.PaginationMeta {
		Page: page,
		Limit: limit,
		Total: int(total),
		TotalPage: int(math.Ceil(float64(total) / float64(limit))),
		Filter: filter,
		Sort: sort,
	}

	return utils.SuccessPagination(ctx, "Data board berhasil diambil", boards, meta)
}