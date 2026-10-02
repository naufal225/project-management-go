package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/services"
	"github.com/naufal225/project-management-go/utils"
)

type ListController struct {
	service services.ListService
}

func NewListController(s services.ListService) *ListController {
	return &ListController{service: s}
}

func (c *ListController) CreateList(ctx *fiber.Ctx) error {
	list := new(models.List)
	if err := ctx.BodyParser(list); err != nil {
		return utils.BadRequest(ctx, "gagal membaca request", err.Error())
	}

	if err := c.service.Create(list); err != nil {
		return utils.BadRequest(ctx, "gagal membuat list", err.Error())
	}

	return utils.Success(ctx, "list berhasil dibuat", list);
}

func (c *ListController) UpdateList(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	list := new(models.List)

	if err := ctx.BodyParser(list); err != nil {
		return utils.BadRequest(ctx, "gagal parsing data", err.Error())
	}

	if _, err := uuid.Parse(publicID); err != nil {
		return utils.BadRequest(ctx, "id tidak valid", err.Error())
	}

	exisitingList, err := c.service.GetByPublicID(publicID)
	if err != nil {
		return  utils.NotFound(ctx, "list tidak ditemukan", err.Error())
	}

	list.InternalID = exisitingList.InternalID
	list.PublicID = exisitingList.PublicID

	if err := c.service.Update(list); err != nil {
		return  utils.BadRequest(ctx, "gagal update list", err.Error())
	}

	updatedList, err := c.service.GetByPublicID(publicID)
	if err != nil {
		return utils.NotFound(ctx, "list tidak ditemukan", err.Error())
	}

	return utils.Success(ctx, "berhasil memperbarui list", updatedList)
}

func (c *ListController) GetListOnBoard(ctx *fiber.Ctx) error {
	boardPublicID := ctx.Params("board_id")

	if _, err := uuid.Parse(boardPublicID); err != nil {
		return utils.BadRequest(ctx, "id tidak valid", err.Error())
	}

	lists, err := c.service.GetByBoardID(boardPublicID)
	if err != nil {
		return utils.NotFound(ctx, "list tidak ditemukan", err.Error())
	}

	return utils.Success(ctx, "data berhasil diambil", lists)
}

func (c *ListController) DeleteList(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id");
	if _, err := uuid.Parse(publicID); err != nil {
		return  utils.BadRequest(ctx, "id tidak valid", err.Error());
	}

	list, err := c.service.GetByPublicID(publicID);
	if err != nil {
		return utils.NotFound(ctx, "list tidak ditemukan", err.Error())
	}

	if err := c.service.Delete(uint(list.InternalID)); err != nil {
		utils.InternalServerError(ctx, "Gagal menghapus list", err.Error())
	}

	return utils.Success(ctx, "List berhasil dihapus", publicID);
}

func (c *ListController) UpdateListPosition(ctx *fiber.Ctx) error {
	boardID := ctx.Params("board_id");
	if _, err := uuid.Parse(boardID); err != nil {
		return utils.BadRequest(ctx, "id tidak valid", err.Error());
	}

	var positionUUID []uuid.UUID
	if err := ctx.BodyParser(&positionUUID); err != nil {
		// jika gagal, coba parse sebagai array of string
		var positionString []string 
		if err := ctx.BodyParser(&positionString); err != nil {
			return utils.BadRequest(ctx, "invalid position format", err.Error())
		}
		// konversi dari string ke uuid
		for _, s := range positionString {
			u, err := uuid.Parse(s)
			if err != nil {
				return utils.BadRequest(ctx, "invalid position format", err.Error())
			}

			positionUUID = append(positionUUID, u)
		}
	}

	if err := c.service.UpdatePosition(boardID, positionUUID); err != nil {
		return utils.InternalServerError(ctx, "gagal update list", err.Error())
	}

	return utils.Success(ctx, "posisi list berhasil diperbarui", nil);
}