package routes

import (
	"github.com/catalinfl/readit-api/controllers"
	"github.com/gofiber/fiber/v2"
)

func booksRoute(api fiber.Router) {
	bookRoute := api.Group("/books")

	bookRoute.Get("/", controllers.GetAllBooks)
	bookRoute.Get("/verify", controllers.VerifyLogin)
	bookRoute.Post("/", controllers.CreateBook)

	bookRoute.Get("/user-books", controllers.GetAllUserBooks)
	bookRoute.Get("/user-books/:id", controllers.GetUserBooks)
	bookRoute.Delete("/user-books/:bookId", controllers.DeleteUserBook)
	bookRoute.Post("/user-books", controllers.CreateUserBook)

	bookRoute.Get("/get-paginated", controllers.GetBooksPaginated)
	bookRoute.Get("/get-infinite/:id", controllers.GetBooks)

	bookRoute.Put("/edit-pages", controllers.UpdateReadingBook)
	bookRoute.Put("/edit-genre", controllers.UpdateGenre)
}
