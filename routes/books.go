package routes

import (
	"github.com/catalinfl/readit-api/handlers"
	"github.com/gofiber/fiber/v2"
)

func booksRoute(api fiber.Router) {
	bookRoute := api.Group("/books")

	bookRoute.Get("/", handlers.GetAllBooks)
	bookRoute.Post("/", handlers.CreateBook)

	bookRoute.Get("/user-books", handlers.GetAllUserBooks)
	bookRoute.Get("/user-books/:id", handlers.GetUserBooks)
	bookRoute.Delete("/user-books/:bookId", handlers.DeleteUserBook)
	bookRoute.Post("/user-books", handlers.CreateUserBook)

	bookRoute.Get("/get-paginated", handlers.GetBooksPaginated)
	bookRoute.Get("/get-infinite/:id", handlers.GetBooks)

	bookRoute.Put("/edit-pages", handlers.UpdateReadingBook)
	bookRoute.Put("/edit-genre", handlers.UpdateGenre)
}
