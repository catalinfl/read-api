package routes

import (
	"github.com/catalinfl/readit-api/handlers"
	"github.com/gofiber/fiber/v2"
)

func usersRoute(api fiber.Router) {
	userRoute := api.Group("/users")

	userRoute.Post("/", handlers.CreateUser)
	userRoute.Get("/", handlers.GetUsers)
	userRoute.Get("/:id", handlers.GetUser)
}
