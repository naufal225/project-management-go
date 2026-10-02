package controllers

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	dtos "github.com/naufal225/project-management-go/DTOs"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/services"
	"github.com/naufal225/project-management-go/utils"
)

type UserController struct {
	services services.UserService
}

func NewUserController(s services.UserService) *UserController {
	return &UserController{services: s}
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
	user := new(models.User)
	req := new(dtos.RegisterRequest)

	if err := ctx.BodyParser(&req); err != nil {
		return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Password = req.Password

	if err := c.services.Register(user); err != nil {
		return utils.BadRequest(ctx, "Registrasi gagal", err.Error())
	}

	var userResponse models.UserResponse

	_ = copier.Copy(&userResponse, &user)

	return utils.Success(ctx, "Register success", userResponse)
}

func (c *UserController) Login(ctx *fiber.Ctx) error {
	req := new(dtos.LoginRequest)

	if err := ctx.BodyParser(&req); err != nil {
		return utils.BadRequest(ctx, "Invalid request", err.Error())
	}

	user, err := c.services.Login(req.Email, req.Password)
	if err != nil {
		return utils.Unauthorized(ctx, "Login failed", err.Error())
	}

	token, _ := utils.GenerateToken(int64(user.InternalID), user.Role, user.Email, user.PublicID)
	refreshToken, _ := utils.GenerateRefreshToken(int64(user.InternalID))

	var userResponse models.UserResponse

	_ = copier.Copy(&userResponse, &user)

	return utils.Success(ctx, "Login success", fiber.Map{
		"access_token":  token,
		"refresh_token": refreshToken,
		"user":          userResponse,
	})
}

func (c *UserController) GetUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	user, err := c.services.GetByPublicID(id)
	if err != nil {
		return utils.NotFound(ctx, "Data Not Found", err.Error())
	}

	var userResponse models.UserResponse
	err = copier.Copy(&userResponse, user)

	if err != nil {
		return utils.BadRequest(ctx, "Internal Parse Error", err.Error())
	}

	return utils.Success(ctx, "Data is found", userResponse)
}

func (c *UserController) GetUserPagination(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset := (page - 1) * limit

	filter := ctx.Query("filter", "")
	sort := ctx.Query("sort", "")

	users, total, err := c.services.GetAllPagination(filter, sort, limit, offset)

	if err != nil {
		return utils.BadRequest(ctx, "Gagal mengambil data", err.Error())
	}

	var usersResponse []models.UserResponse
	err = copier.Copy(&usersResponse, &users)

	if err != nil {
		return utils.BadRequest(ctx, "Internal Parse Error", err.Error())
	}

	meta := utils.PaginationMeta {
		Page: page,
		Limit: limit,
		Total: int(total),
		TotalPage: int(math.Ceil(float64(total)/float64(limit))),
		Filter: filter,
		Sort: sort,
	}

	if total == 0 {
		return utils.NotFoundPagination(ctx, "Data pengguna tidak ditemukan", usersResponse, meta)
	}

	return  utils.SuccessPagination(ctx, "Data ditemukan", usersResponse, meta)
}

func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	publicID, err := uuid.Parse(id)

	if err != nil {
		return utils.BadRequest(ctx, "Invalid ID Format", err.Error())
	}

	authPublicID, authRole, err := utils.GetUserClaims(ctx)

	fmt.Println("role:", authRole)

	if err != nil {
		return utils.Unauthorized(ctx, "Anda tidak memiliki akses untuk melakukan tindakan ini", err.Error())
	}

	if authRole != "admin" && authPublicID != publicID {
		return utils.Forbidden(ctx, "Anda tidak memiliki akses untuk melakukan tindakan ini", "Forbidden")
	}
	
	userUpdated, err := c.services.GetByPublicID(id)
	if err != nil {
		return utils.BadRequest(ctx, "User tidak ditemukan", err.Error())
	}

	if authRole == "admin" {
		var updateUserReq dtos.AdminUpdateUserRequest
		if err := ctx.BodyParser(&updateUserReq); err != nil {
			return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
		}

		userUpdated.Name = updateUserReq.Name
		userUpdated.Email = updateUserReq.Email
		if updateUserReq.Role != "" {
			userUpdated.Role = updateUserReq.Role
		}
	} else {
		var updateUserReq dtos.UpdateUserRequest
		if err := ctx.BodyParser(&updateUserReq); err != nil {
			return utils.BadRequest(ctx, "Gagal parsing data", err.Error())
		}
		userUpdated.Name = updateUserReq.Name
		userUpdated.Email = updateUserReq.Email
	}

	userUpdated.PublicID = publicID

	if err := c.services.Update(userUpdated); err != nil {
		return utils.BadRequest(ctx, "Gagal mengupdate data user", err.Error())
	}

	var userResponse models.UserResponse
	err = copier.Copy(&userResponse, &userUpdated)
	if err != nil {
		return utils.InternalServerError(ctx, "Internal Parse Error", err.Error())
	}

	return utils.Success(ctx, "Berhasil update data", userResponse)
}

func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	_, authRole, err := utils.GetUserClaims(ctx)
	if err != nil {
		return utils.Unauthorized(ctx, "Anda tidak memiliki akses untuk melakukan tindakan ini", err.Error())
	}

	if authRole != "admin" {
		return utils.Forbidden(ctx, "Anda tidak memiliki akses untuk melakukan tindakan ini", "Forbidden")
	}

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return  utils.BadRequest(ctx, "ID tidak valid", err.Error())
	}	

	if err := c.services.Delete(id); err != nil {
		return utils.InternalServerError(ctx, "Gagal menghapus data", err.Error())
	}

	return utils.Success(ctx, "Berhasil menghapus data", id)
}