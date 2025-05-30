package routes

import (
	"backend-pedika-fiber/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetAuthRoutes(app *fiber.App) {
	userGroup := app.Group("/api/user")
	{
		userGroup.Post("/register", handlers.RegisterUser)
		userGroup.Post("/login", handlers.LoginUser)
	}
}
