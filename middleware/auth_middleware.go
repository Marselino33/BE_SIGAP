package middleware

import (
	"backend-pedika-fiber/helper"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	log.Printf("User  request: %s %s", c.Method(), c.Path())
	authHeader := c.Get("Authorization")
	log.Printf("Authorization header: %s", authHeader)
	if authHeader == "" {
		response := helper.ResponseWithOutData{
			Code:    fiber.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Missing token",
		}
		return c.Status(fiber.StatusUnauthorized).JSON(response)
	}
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		response := helper.ResponseWithOutData{
			Code:    fiber.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token format",
		}
		return c.Status(fiber.StatusUnauthorized).JSON(response)
	}

	tokenString := splitToken[1]
	log.Printf("Token string extracted: %s", tokenString) // Log the extracted token string
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		response := helper.ResponseWithOutData{
			Code:    fiber.StatusUnauthorized,
			Status:  "error",
			Message: "Unauthorized: Invalid token",
		}
		return c.Status(fiber.StatusUnauthorized).JSON(response)
	}
	fmt.Println("Middleware With Auth Middleware")
	c.Locals("user", token)
	return c.Next()
}
