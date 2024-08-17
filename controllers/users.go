package controllers

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/middlewares"
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
	claims["id"] = user.ID

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

func BookStructToMap(obj interface{}) map[string]interface{} {
	val := reflect.ValueOf(obj)

	typ := reflect.TypeOf(obj)

	data := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		if field.Kind() == reflect.Struct {
			continue
		}

		fieldName := strings.ToLower(typ.Field(i).Name)

		if fieldName == "bookid" || fieldName == "userid" || fieldName == "userbooksid" || fieldName == "users" {
			continue
		}

		data[fieldName] = field.Interface()

	}

	return data
}

func SendFriendRequest(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	num, err := strconv.Atoi(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	if int(usr["id"].(float64)) == num {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	var friendRequestReceived models.Friends

	rows := db.GetDB().Where("receiver_id = ? AND sender_id = ?", num, int(usr["id"].(float64))).First(&friendRequestReceived)

	if rows.RowsAffected > 0 {
		return c.Status(400).JSON(fiber.Map{
			"message": "Friend request already sent",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	var friendRequest models.Friends

	friendRequest.SenderID = int(usr["id"].(float64))
	friendRequest.SenderName = usr["name"].(string)
	friendRequest.ReceiverName = user.Name
	friendRequest.ReceiverID = int(user.ID)
	friendRequest.Status = "pending"

	db.GetDB().Create(&friendRequest)

	return c.Status(200).JSON(fiber.Map{
		"message": "Friend request sent",
		"req":     friendRequest,
	})
}

func AcceptFriendRequest(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var friendRequest models.Friends

	db.GetDB().Where("id = ?", id).First(&friendRequest)

	if friendRequest.SenderID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "Friend request not found",
		})
	}

	if int(usr["id"].(float64)) != friendRequest.ReceiverID {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	friendRequest.Status = "accepted"

	db.GetDB().Save(&friendRequest)

	return c.Status(200).JSON(fiber.Map{
		"message": "Friend request accepted",
	})
}

func GetFriendRequests(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var friendRequests []models.Friends

	db.GetDB().Where("receiver_id = ?", int(usr["id"].(float64))).Find(&friendRequests)

	var response []map[string]interface{}

	var t map[string]interface{}

	for _, req := range friendRequests {
		t = make(map[string]interface{})

		t["id"] = req.SenderID
		t["name"] = req.SenderName

		response = append(response, t)
	}

	return c.JSON(fiber.Map{
		"message": response,
	})
}

func GetAllFriendsRequests(c *fiber.Ctx) error {

	var friendRequests []models.Friends

	db.GetDB().Find(&friendRequests)

	return c.JSON(fiber.Map{
		"message": friendRequests,
	})
}

func DeleteFriendRequest(c *fiber.Ctx) error {
	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var friendRequest models.Friends

	// delete friend request

	db.GetDB().Where("sender_id = ? AND receiver_id = ?", int(usr["id"].(float64)), request["id"]).First(&friendRequest)

	if friendRequest.SenderID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "Friend request not found",
		})
	}

	db.GetDB().Delete(&friendRequest)

	return c.Status(200).JSON(fiber.Map{
		"message": "Friend request deleted",
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

	response := make(map[string]interface{})

	response["name"] = user.Name
	response["email"] = user.Email
	response["rank"] = user.Rank
	response["librarian"] = user.Librarian
	response["admin"] = user.Admin
	response["profile_pic"] = user.ProfilePic
	response["real_name"] = user.RealName

	db.GetDB().Model(&user).Association("Books").Find(&user.Books)

	var userBooks []models.UserBooks

	db.GetDB().Where("user_id = ?", user.ID).Find(&userBooks)

	var responseBooks []map[string]interface{}

	for _, booksFromDb := range userBooks {
		for _, booksFromUserRes := range user.Books {
			if booksFromDb.BookID == uint(booksFromUserRes.ID) {
				bookFromDbMap := BookStructToMap(booksFromDb)
				bookFromUserResMap := BookStructToMap(booksFromUserRes)

				for k, v := range bookFromDbMap {
					bookFromUserResMap[k] = v
				}

				responseBooks = append(responseBooks, bookFromUserResMap)
			}
		}
	}

	response["books"] = responseBooks

	return c.JSON(fiber.Map{
		"message": response,
	})

}

func DeleteUser(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	db.GetDB().Delete(&user)

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})

}

func PromoteToLibrarian(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"message": "Invalid request",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	user.Librarian = true

	db.GetDB().Save(&user)

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("User %s has been promoted to librarian", user.Name),
	})

}

func ModifyUser(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"message": "Error at parsing json",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	for key, val := range request {
		switch key {
		case "name":
			user.Name = val.(string)
		case "email":
			user.Email = val.(string)
		case "password":
			password := hashPassword(val.(string))
			user.Password = password
		case "rank":
			user.Rank = val.(string)
		}
	}

	db.GetDB().Save(&user)

	return c.Status(200).JSON(fiber.Map{
		"message": "User has been modified",
	})

}
