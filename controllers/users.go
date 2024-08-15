package controllers

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
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
		"message": "User created successfully",
	})
}

func hashPassword(password string) string {
	cryptPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	return string(cryptPass)
}

func Login(c *fiber.Ctx) error {

	godotenv.Load()

	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	var user models.User

	fmt.Println(request["name"])

	db.GetDB().Where("name = ?", request["name"]).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	if !checkPassword(user.Password, request["password"].(string)) {
		return c.Status(400).JSON(fiber.Map{
			"message": "Incorrect password",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = user.Name

	secret_jwt := os.Getenv("JWT_TOKEN_SECRET")

	if secret_jwt == "" {
		fmt.Println("Error loading JWT Key")
	}

	tok, err := token.SignedString([]byte(secret_jwt))

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Could not login",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    tok,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		SameSite: "Strict",
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "Login successful",
	})

}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	return err == nil
}

func GetUsers(c *fiber.Ctx) error {

	var users []models.User

	db.GetDB().Find(&users)

	return c.JSON(fiber.Map{
		"message": users,
	})

}

func VerifyLogin(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	fmt.Println(token)

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	str := verifyToken(token)

	if str == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized, please log in",
		})
	}

	c.Next()

	return nil
}

func verifyToken(token string) string {

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
		return claims["name"].(string)
	}

	return ""

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
