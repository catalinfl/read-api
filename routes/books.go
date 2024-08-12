package routes

import (
	"github.com/catalinfl/readit-api/handlers"
	"github.com/gofiber/fiber/v2"
)

func booksRoute(api fiber.Router) {
	bookRoute := api.Group("/books")

	bookRoute.Get("/", handlers.GetAllBooks)
	bookRoute.Post("/user-books", handlers.CreateUserBook)
	bookRoute.Get("/user-books", handlers.GetAllUserBooks)
	bookRoute.Get("/user-books/:id", handlers.GetUserBooks)
	bookRoute.Get("/get-paginated", handlers.GetBooksPaginated)
	bookRoute.Get("/get-infinite/:id", handlers.GetBooks)
	bookRoute.Post("/", handlers.CreateBook)
}
