package middlewares

import (
	"fmt"
	"os"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
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

func VerifyLogin(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, please log in",
		})
	}

	str := VerifyTokenAndParse(token)

	if str == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, something happened with jwt token",
		})
	}

	return c.Next()

}

func VerifyIfLibrarian(c *fiber.Ctx) error {
	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, you are not a librarian",
		})
	}

	str := VerifyTokenAndParse(token)

	if str == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, you are not a librarian",
		})
	}

	var user models.User

	db.GetDB().Where("name = ?", str).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if !user.Librarian {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, you are not a librarian",
		})
	}

	return c.Next()

}

func VerifyIfAdmin(c *fiber.Ctx) error {
	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	str := VerifyTokenAndParse(token)

	if str == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var user models.User

	db.GetDB().Where("name = ?", str).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if !user.Admin {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Next()
}

func CountUserBooks(c *fiber.Ctx) error {
	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	str := VerifyTokenAndParse(token)

	if str == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, please log in",
		})
	}

	var user models.User

	db.GetDB().Where("name = ?", str).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	var userBooks models.UserBooks

	var userBooksLength int64 = 0

	db.GetDB().Where("user_id = ?", user.ID).Find(&userBooks).Count(&userBooksLength)

	if userBooksLength < 10 && user.Rank != "Bronze" {
		user.Rank = "Bronze"
		db.GetDB().Model(&user).Updates(user)
		c.Next()

		return nil
	}

	if (userBooksLength > 10 && userBooksLength < 20) && user.Rank != "" {
		user.Rank = "Silver"
		db.GetDB().Model(&user).Updates(user)
		c.Next()
		return nil
	}

	if userBooksLength > 20 && user.Rank != "Gold" {
		user.Rank = "Gold"
		db.GetDB().Model(&user).Updates(user)
		c.Next()
		return nil
	}

	return c.Next()

}

func VerifyTokenAndParse(token string) map[string]interface{} {

	godotenv.Load()

	secret_jwt := os.Getenv("JWT_TOKEN_SECRET")

	if secret_jwt == "" {
		fmt.Println("Error loading JWT Key")
	}

	tok, err := jwt.Parse(token, func(jwt *jwt.Token) (interface{}, error) {
		return []byte(secret_jwt), nil
	})

	if err != nil {
		fmt.Println(err)
	}

	if !tok.Valid {
		fmt.Println("Token is not valid")
	}

	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
		return claims
	}

	return nil

}
