package main

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/migration"
	"backend-pedika-fiber/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	database.GetDBInstance()
	migration.RunMigration()
	routes.SetAuthRoutes(app)
	routes.SetAdminRoutes(app)
	routes.SetMasyarakatRoutes(app)
	routes.RoutesWithOutLogin(app)
	app.Listen(":8080")
}
