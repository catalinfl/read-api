package routes

import (
	"github.com/catalinfl/readit-api/controllers"
	"github.com/catalinfl/readit-api/middlewares"
	"github.com/gofiber/fiber/v2"
)

func usersRoute(api fiber.Router) {
	userRoute := api.Group("/users")

	userRoute.Post("/", controllers.CreateUser)
	userRoute.Post("/login", controllers.Login)
	userRoute.Get("/:id", middlewares.VerifyLogin, controllers.GetUser)
}
