package handlers

import (
	"regexp"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *fiber.Ctx) error {

	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	if len(user.Name) < 3 && len(user.Name) > 100 {
		return c.Status(400).JSON(fiber.Map{
			"message": "Name must be between 3 and 100 characters",
		})
	}

	if len(user.Email) < 3 && len(user.Email) > 100 {
		return c.Status(400).JSON(fiber.Map{
			"message": "Email must be between 3 and 100 characters",
		})
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if !emailRegex.MatchString(user.Email) {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid email address",
		})
	}

	letterRegex := regexp.MustCompile(`[A-Za-z]`)
	digitRegex := regexp.MustCompile(`[0-9]`)

	if !letterRegex.MatchString(user.Password) || !digitRegex.MatchString(user.Password) || len(user.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{
			"message": "Password must contain at least 8 characters, one letter and one digit",
		})
	}

	user.Password = hashPassword(user.Password)

	var existingUser models.User

	db.GetDB().Where("name = ?", user.Name).First(&existingUser)

	if existingUser.ID > 0 {
		return c.Status(400).JSON(fiber.Map{
			"message": "User already exists",
		})
	}

	db.GetDB().Create(&user)

	return c.JSON(fiber.Map{
		"message": user,
	})
}

func hashPassword(password string) string {
	cryptPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(cryptPass)
}

func GetUsers(c *fiber.Ctx) error {

	var users []models.User

	db.GetDB().Find(&users)

	return c.JSON(fiber.Map{
		"message": users,
	})

}

func GetUser(c *fiber.Ctx) error {

	var user models.User

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": user,
	})

}
