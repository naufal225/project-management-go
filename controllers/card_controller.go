package controllers

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	dtos "github.com/naufal225/project-management-go/DTOs"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/services"
	"github.com/naufal225/project-management-go/utils"
)

type CardController struct {
	cardService       services.CardService
	attachmentService services.AttachmentService
	userService       services.UserService
	uploadDir         string
}

func NewCardController(cardService services.CardService, attachmentService services.AttachmentService, userService services.UserService, uploadDir string) *CardController {
	return &CardController{
		cardService:       cardService,
		attachmentService: attachmentService,
		userService:       userService,
		uploadDir:         uploadDir,
	}
}

func (c *CardController) CreateCard(ctx *fiber.Ctx) error {
	var req dtos.CreateCardRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.BadRequest(ctx, "Gagal mengambil data", err.Error())
	}

	card := &models.Card{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     &req.DueDate,
		Position:    req.Position,
	}

	if err := c.cardService.Create(card, req.ListPublicID); err != nil {
		return utils.InternalServerError(ctx, "Gagal membuat card", err.Error())
	}

	return utils.Success(ctx, "Berhasil membuat card", card)
}

func (c *CardController) UpdateCard(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")

	var req dtos.UpdateCardRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
	}

	if _, err := uuid.Parse(publicID); err != nil {
		return utils.BadRequest(ctx, "Id tidak valid", err.Error())
	}

	card := &models.Card{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Position:    req.Position,
		PublicID:    uuid.MustParse(publicID),
	}

	if err := c.cardService.Update(card, req.ListPublicID); err != nil {
		return utils.InternalServerError(ctx, "Gagal update data", err.Error())
	}

	return utils.Success(ctx, "Berhasil update card", card)
}

func (c *CardController) DeleteCard(ctx *fiber.Ctx) error {
	publicID := ctx.Params("id")
	if _, err := uuid.Parse(publicID); err != nil {
		return utils.BadRequest(ctx, "ID tidak valid", err.Error())
	}

	card, err := c.cardService.GetByPublicID(publicID)
	if err != nil {
		return utils.NotFound(ctx, "card tidak ditemukan", err.Error())
	}

	if err := c.cardService.Delete(uint(card.InternalID)); err != nil {
		return utils.BadRequest(ctx, "Gagal menghapus data", err.Error())
	}

	return utils.Success(ctx, "card berhasil dihapus", publicID)
}

func (c *CardController) GetCardDetail(ctx *fiber.Ctx) error {
	cardPublicID := ctx.Params("id")
	if _, err := uuid.Parse(cardPublicID); err != nil {
		return utils.BadRequest(ctx, "ID tidak valid", err.Error())
	}

	card, err := c.cardService.GetByPublicID(cardPublicID)
	if err != nil {
		return utils.InternalServerError(ctx, "Error saat mengambil data", err.Error())
	}

	if card == nil {
		return utils.NotFound(ctx, "Card tidak ditemukan", "error")
	}

	return utils.Success(ctx, "Card berhasil diambil", card)
}

func (c *CardController) AddCardLabel(ctx *fiber.Ctx) error {
	cardID := ctx.Params("id")

	var body struct {
		LabelID string `json:"label_id"`
	}

	if err := ctx.BodyParser(&body); err != nil {
		return utils.BadRequest(ctx, "Id tidak valid", err.Error())
	}

	if err := c.cardService.AddLabel(cardID, body.LabelID); err != nil {
		return utils.BadRequest(ctx, "Gagal menambahkan data label", err.Error())
	}

	return utils.Success(ctx, "berhasil menambahkan label", nil)
}

func (c *CardController) RemoveCardLabel(ctx *fiber.Ctx) error {
	cardID := ctx.Params("id")

	var body struct {
		LabelID string `json:"label_id"`
	}

	if err := ctx.BodyParser(&body); err != nil {
		return utils.BadRequest(ctx, "Id tidak valid", err.Error())
	}

	if err := c.cardService.RemoveLabel(cardID, body.LabelID); err != nil {
		return utils.BadRequest(ctx, "Gagal menambahkan data label", err.Error())
	}

	return utils.Success(ctx, "berhasil manghapus label", nil)
}

func (c *CardController) UploadAttachment(ctx *fiber.Ctx) error {
	cardPublicID := ctx.Params("id")
	if cardPublicID == "" {
		return utils.BadRequest(ctx, "card id is required", "error")
	}

	u := ctx.Locals("user")
	if u == nil {
		return utils.Unauthorized(ctx, "Unauthorize", "error unauthorize")
	}

	token, ok := u.(*jwt.Token)
	if !ok {
		return utils.Unauthorized(ctx, "Unauthorize", "error unauthorize")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return utils.Unauthorized(ctx, "Unauthorize", "error unauthorize")
	}

	userPublicID, _ := claims["pub_id"].(string)
	if userPublicID == "" {
		return utils.Unauthorized(ctx, "Unauthorize: user id not found", "error unauthorize")
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return utils.BadRequest(ctx, "File is required", err.Error())
	}

	const maxSize = 20 << 20
	if fileHeader.Size > maxSize {
		return utils.BadRequest(ctx, "File too large (max 20 MB)", "Error")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}

	if !allowed[ext] {
		return utils.BadRequest(ctx, "file type not allowed", "error")
	}

	if err := os.MkdirAll(c.uploadDir, os.ModePerm); err != nil {
		return utils.InternalServerError(ctx, "cannot create upload directory", "Error")
	}

	dstFileName := uuid.New().String() + ext
	dstPath := filepath.Join(c.uploadDir, dstFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return utils.InternalServerError(ctx, "cannot open uploaded file", "Error")
	}

	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return utils.InternalServerError(ctx, "cannot create destination file", err.Error())
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return utils.InternalServerError(ctx, "cannot save file", err.Error())
	}

	attachment, err := c.attachmentService.Create(cardPublicID, userPublicID, dstPath)
	if err != nil {
		_ = os.Remove(dstPath)
		return utils.InternalServerError(ctx, err.Error(), "Error")
	}

	return ctx.Status(fiber.StatusCreated).JSON(attachment)
}

func (c *CardController) GetAttachment(ctx *fiber.Ctx) error {
	cardPublicID := ctx.Params("id")
	attachments, err := c.attachmentService.GetByCardID(cardPublicID)
	if err != nil {
		return utils.InternalServerError(ctx, err.Error(), "Error")
	}

	return ctx.JSON(attachments)
}

func (c *CardController) DeleteAttachment(ctx *fiber.Ctx) error {
	pubIDstr := ctx.Params("attachment_id")
	pubID, err := uuid.Parse(pubIDstr)
	if err != nil {
		return utils.BadRequest(ctx, "invalid attachment id", err.Error())
	}

	attachments, err := c.attachmentService.GetByPublicID(pubID)
	if err != nil {
		return utils.InternalServerError(ctx, err.Error(), "error")
	}

	if attachments != nil {
		return utils.NotFound(ctx, "Attachment not found", err.Error())
	}

	user := ctx.Locals("user")
	token, _ := user.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return utils.Unauthorized(ctx, "Invalid user", err.Error())
	}

	userID := int64(userIDFloat)

	if attachments.UserID != userID {
		return utils.Unauthorized(ctx, "You don't have permission", "Error")
	}

	if err := c.attachmentService.DeleteByPublicID(pubID); err != nil {
		return utils.InternalServerError(ctx, err.Error(), "Error")
	}

	_ = os.Remove(attachments.File)

	return utils.Success(ctx, "Berhasil dihapus", uuid.Nil)
}

func (c *CardController) AddCardAssignee(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return utils.BadRequest(ctx, "card id tidak valid", err.Error())
	}

	card, err := c.cardService.GetByPublicID(id)
	if err != nil {
		return utils.NotFound(ctx, "Card tidak ditemukan", err.Error())
	}

	var usersIDstr []string
	if err := ctx.BodyParser(&usersIDstr); err != nil {
		return utils.BadRequest(ctx, "user id tidak valid", err.Error())
	}

	var usersID []uint
	for _, uid := range usersIDstr {
		user, err := c.userService.GetByPublicID(uid)
		if err != nil {
			return utils.NotFound(ctx, "user tidak ditemukan", err.Error())
		}
		usersID = append(usersID, uint(user.InternalID))
	}

	if err = c.cardService.AddAssignees(uint(card.InternalID), usersID); err != nil {
		return utils.InternalServerError(ctx, "Gagal assign user ke card", err.Error());
	}

	return utils.Success(ctx, "Berhasil menambahkan assignee card ke user", nil)
}

func (c *CardController) RemoveCardAssignee(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return utils.BadRequest(ctx, "card id tidak valid", err.Error())
	}

	card, err := c.cardService.GetByPublicID(id)
	if err != nil {
		return utils.NotFound(ctx, "card tidak ditemukan", err.Error())
	}

	var userIDstr string
	if err := ctx.BodyParser(&userIDstr); err != nil {
		return utils.BadRequest(ctx, "id user tidak valid", err.Error())
	}

	user, err := c.userService.GetByPublicID(string(userIDstr))
	if err != nil {
		return utils.NotFound(ctx, "user tidak ditemukan", err.Error())
	}

	if err :=  c.cardService.RemoveAssignees(uint(card.InternalID), uint(user.InternalID)); err != nil {
		return  utils.InternalServerError(ctx, "gagal menghapus assignee card ke user", err.Error())
	}

	return utils.Success(ctx, "Berhasil menghapus assignee card ke user", userIDstr)
}
