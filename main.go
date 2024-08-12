package main

import (
	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	db.Connect()

	routes.Setup(app)

	app.Listen(":3000")

}
