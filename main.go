package main

import (
	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/middlewares"
	"github.com/catalinfl/readit-api/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	db.Connect()

	app.Use(middlewares.UseCORS())

	routes.Setup(app)

	app.Listen(":3000")

}
