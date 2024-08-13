package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

func UseCORS() any {
	c := cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, DELETE",
	})

	fmt.Println("CORS middleware enabled")

	return c
}
