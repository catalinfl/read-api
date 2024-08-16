package routes

import (
	"github.com/catalinfl/readit-api/controllers"
	"github.com/gofiber/fiber/v2"
)

func adminRoute(api fiber.Router) {

	adminRoute := api.Group("/admin")

	// adminRoute.Use(middlewares.VerifyIfAdmin)

	adminRoute.Get("/users", controllers.GetUsers)
	adminRoute.Put("/promote/:id", controllers.PromoteToLibrarian)
	adminRoute.Put("/users/:id", controllers.ModifyUser)
	adminRoute.Delete("/users/:id", controllers.DeleteUser)
	adminRoute.Delete("/book/:id", controllers.DeleteBook)
}
