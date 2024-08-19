package routes

import (
	"github.com/catalinfl/readit-api/controllers"
	"github.com/catalinfl/readit-api/middlewares"
	"github.com/gofiber/fiber/v2"
)

func librarianRoute(api fiber.Router) {
	librarianRoute := api.Group("/librarian")

	librarianRoute.Use(middlewares.VerifyIfLibrarian)

	librarianRoute.Post("/create-book", controllers.CreateBook)
	librarianRoute.Put("/modify-book/:id", controllers.ModifyBook)
	librarianRoute.Delete("/delete-book/:bookId", controllers.DeleteBookLibrarian)

}
