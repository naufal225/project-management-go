package utils

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/naufal225/project-management-go/config"
)

// generate token
//generate refresh token
func GenerateToken(userID int64, role, email string, publicID uuid.UUID) (string, error) {
	secret := config.AppConfig.JWTSecret
	duration, _ := time.ParseDuration(config.AppConfig.JWTExpire)

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"pub_id": publicID.String(),
		"email": email,
		"exp": time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

//refresh token
func GenerateRefreshToken(userID int64) (string, error) {
	secret := config.AppConfig.JWTSecret
	duration, _ := time.ParseDuration(config.AppConfig.JWTRefreshToken)

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp": time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

// getUserClaims
func GetUserClaims(ctx *fiber.Ctx) (uuid.UUID, string, error) {
	userToken := ctx.Locals("user")
	if userToken == nil {
		return uuid.Nil, "", fiber.NewError(fiber.StatusUnauthorized, "Token tidak valid atau hilang")
	}

	var claims jwt.MapClaims
	
	if token, ok := userToken.(*jwt.Token); ok {
		if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
			claims = mapClaims
		}
	} else if mapClaims, ok := userToken.(jwt.MapClaims); ok {
		claims = mapClaims
	}

	if claims == nil {
		return uuid.Nil, "", fiber.NewError(fiber.StatusUnauthorized, "Token tidak valid atau hilang")
	}

	publicIDRaw, exists := claims["pub_id"]
	if !exists || publicIDRaw == nil {
		return uuid.Nil, "", fiber.NewError(fiber.StatusUnauthorized, "Token tidak valid atau hilang")
	}

	publicIDStr := fmt.Sprintf("%v", publicIDRaw)
	publicID, err := uuid.Parse(publicIDStr)
	if err != nil {
		return uuid.Nil, "", fiber.NewError(fiber.StatusUnauthorized, "Token tidak valid atau hilang")
	}

	role, _ := claims["role"].(string)
	return publicID, role, nil
}
