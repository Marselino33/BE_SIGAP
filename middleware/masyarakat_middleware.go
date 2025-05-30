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


func MasyarakatMiddleware(c *fiber.Ctx) error {
    authHeader := c.Get("Authorization")
    log.Printf("Authorization header: %s", authHeader) // Log the authorization header

    if authHeader == "" {
        response := helper.ResponseWithOutData{
            Code:    fiber.StatusUnauthorized,
            Status:  "error",
            Message: "Unauthorized: Missing token",
        }
        log.Println("Unauthorized: Missing token") // Log the error
        return c.Status(fiber.StatusUnauthorized).JSON(response)
    }

    splitToken := strings.Split(authHeader, "Bearer ")
    if len(splitToken) != 2 {
        response := helper.ResponseWithOutData{
            Code:    fiber.StatusUnauthorized,
            Status:  "error",
            Message: "Unauthorized: Invalid token format",
        }
        log.Println("Unauthorized: Invalid token format") // Log the error
        return c.Status(fiber.StatusUnauthorized).JSON(response)
    }
    tokenString := splitToken[1]
    log.Printf("Token string extracted: %s", tokenString) // Log the extracted token string

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(os.Getenv("JWT_SECRET_KEY")), nil
    })
    if err != nil || !token.Valid {
        response := helper.ResponseWithOutData{
            Code:    fiber.StatusUnauthorized,
            Status:  "error",
            Message: "Unauthorized: Invalid or expired token",
        }
        log.Println("Unauthorized: Invalid or expired token") // Log the error
        return c.Status(fiber.StatusUnauthorized).JSON(response)
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        response := helper.ResponseWithOutData{
            Code:    fiber.StatusUnauthorized,
            Status:  "error",
            Message: "Unauthorized: Invalid token claims",
        }
        log.Println("Unauthorized: Invalid token claims") // Log the error
        return c.Status(fiber.StatusUnauthorized).JSON(response)
    }

    role, ok := claims["role"].(string)
    if !ok || role != "masyarakat" {
        response := helper.ResponseWithOutData{
            Code:    fiber.StatusForbidden,
            Status:  "error",
            Message: "Forbidden: Only masyarakat can access this endpoint",
        }
        log.Println("Forbidden: Only masyarakat can access this endpoint") // Log the error
        return c.Status(fiber.StatusForbidden).JSON(response)
    }

    // Simpan token ke locals
    c.Locals("user", token)
    log.Printf("Token successfully validated for user: %v", claims["user_id"]) // Log successful validation
    return c.Next()
}
