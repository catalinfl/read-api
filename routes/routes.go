package routes

import (
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	booksRoute(api)
	usersRoute(api)
	adminRoute(api)
	librarianRoute(api)

}
