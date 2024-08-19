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
	userRoute.Get("/serve-photo", controllers.ServePhoto)
	userRoute.Put("/edit-photo", middlewares.VerifyLogin, controllers.EditPhoto)
	userRoute.Delete("/delete-photo", middlewares.VerifyLogin, controllers.DeletePhoto)

	userRoute.Get("/friends", middlewares.VerifyLogin, controllers.GetFriendRequests)
	userRoute.Get("/logout", middlewares.VerifyLogin, controllers.Logout)

	userRoute.Post("/send-friend-request/:id", middlewares.VerifyLogin, controllers.SendFriendRequest)
	userRoute.Put("/accept-friend-request/:id", middlewares.VerifyLogin, controllers.AcceptFriendRequest)
	userRoute.Delete("/reject-friend-request/:id", middlewares.VerifyLogin, controllers.RejectFriendRequest)
	userRoute.Get("/:id", middlewares.VerifyLogin, controllers.GetUser)

}
