package controllers

import (
	"bytes"
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
	"github.com/catalinfl/readit-api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *fiber.Ctx) error {

	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	if len(user.Name) < 3 && len(user.Name) > 100 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Name must be between 3 and 100 characters",
		})
	}

	if len(user.Email) < 3 && len(user.Email) > 100 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Email must be between 3 and 100 characters",
		})
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if !emailRegex.MatchString(user.Email) {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid email address",
		})
	}

	letterRegex := regexp.MustCompile(`[A-Za-z]`)
	digitRegex := regexp.MustCompile(`[0-9]`)

	if !letterRegex.MatchString(user.Password) || !digitRegex.MatchString(user.Password) || len(user.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Password must contain at least 8 characters, one letter and one digit",
		})
	}

	user.Password = hashPassword(user.Password)

	var existingUser models.User

	db.GetDB().Where("name = ? OR email = ?", user.Name, user.Email).First(&existingUser)

	if existingUser.ID > 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "User already exists",
		})
	}

	db.GetDB().Create(&user)

	return c.JSON(fiber.Map{
		"data": "User created successfully",
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
			"data": "Invalid request",
		})
	}

	var user models.User

	db.GetDB().Where("name = ?", request["name"]).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	if !checkPassword(user.Password, request["password"].(string)) {
		return c.Status(400).JSON(fiber.Map{
			"data": "Incorrect password",
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
			"data": "Could not login",
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
		"data": "Login successful",
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
		"data": users,
	})

}

func Logout(c *fiber.Ctx) error {
	c.ClearCookie("jwt_token")

	return c.JSON(fiber.Map{
		"data": "Logged out",
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
			"data": "Invalid request",
		})
	}

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	num, err := strconv.Atoi(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	if int(usr["id"].(float64)) == num {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var friendRequestReceived models.Friends

	rows := db.GetDB().Where("(receiver_id = ? AND sender_id = ?) OR (sender_id = ? AND receiver_id = ?)", int(usr["id"].(float64)), num, int(usr["id"].(float64)), num).First(&friendRequestReceived)

	if rows.RowsAffected > 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "A friend request is already pending, or you are already friend with this user",
		})
	}

	rows = db.GetDB().Where("receiver_id = ? AND sender_id = ?", num, int(usr["id"].(float64))).First(&friendRequestReceived)

	if rows.RowsAffected > 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Friend request already sent",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
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
		"data": "Friend request sent",
		"req":  friendRequest,
	})
}

func EditPhoto(c *fiber.Ctx) error {
	token := c.Cookies("jwt_token")

	t := middlewares.VerifyTokenAndParse(token)

	if t == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized, please log in",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", int(t["id"].(float64))).First(&user)

	if user.ID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	multipartForm, err := c.MultipartForm()

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	files := multipartForm.File["profile_pic"]

	if len(files) != 1 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	file := files[0]

	fileType := file.Header.Get("Content-Type")

	if fileType != "image/jpeg" && fileType != "image/png" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid file type",
		})
	}

	dirPath := "/app/assets/"
	filePath := dirPath + strconv.Itoa(int(user.ID))

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Failed to create directory",
			})
		}
	}

	if file.Size > 1<<20 {
		photoImage, err := utils.CompressPhoto(c, file)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Photo can't be compressed",
			})
		}

		var buf bytes.Buffer

		_, err = utils.ImageToMultipartFileHeader(photoImage, filePath, fileType, &buf)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Photo can't be converted",
			})
		}

		outFile, err := os.Create(filePath)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Photo can't be saved",
			})
		}

		defer outFile.Close()

		if _, err := outFile.Write(buf.Bytes()); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Photo can't be saved",
			})
		}

	} else {
		err := c.SaveFile(file, filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"data": "Photo can't be uploaded",
			})
		}
	}

	// Update the user's profile picture path
	user.ProfilePic = filePath
	db.GetDB().Save(&user)

	return c.Status(200).JSON(fiber.Map{
		"data": "Photo uploaded successfully",
	})
}

func DeletePhoto(c *fiber.Ctx) error {
	token := c.Cookies("jwt_token")

	t := middlewares.VerifyTokenAndParse(token)

	if t == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized, please log in",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", int(t["id"].(float64))).First(&user)

	if user.ID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	if user.ProfilePic == "" {
		return c.Status(404).JSON(fiber.Map{
			"data": "Photo not found",
		})
	}

	err := os.Remove("/app/assets/" + strconv.Itoa(int(user.ID)))

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data": "Failed to delete photo",
		})
	}

	user.ProfilePic = ""

	db.GetDB().Save(&user)

	return c.Status(200).JSON(fiber.Map{
		"data": "Photo deleted successfully",
	})

}

func AcceptFriendRequest(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	var friendRequest models.Friends

	numId, err := strconv.Atoi(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	db.GetDB().Where("(receiver_id = ? AND sender_id = ?) AND status = ?", int(usr["id"].(float64)), numId, "pending").First(&friendRequest)

	if friendRequest.SenderID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Friend request not found",
		})
	}

	if friendRequest.Status == "accepted" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Friend request already accepted",
		})
	}

	if int(usr["id"].(float64)) != friendRequest.ReceiverID {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	friendRequest.Status = "accepted"

	db.GetDB().Where("sender_id = ? AND receiver_id = ?", friendRequest.SenderID, friendRequest.ReceiverID).Save(&friendRequest)

	return c.Status(200).JSON(fiber.Map{
		"data": "Friend request accepted",
	})
}

func GetFriendRequests(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
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
		"data": response,
	})
}

func GetAllFriendsRequests(c *fiber.Ctx) error {

	var friendRequests []models.Friends

	db.GetDB().Find(&friendRequests)

	return c.JSON(fiber.Map{
		"data": friendRequests,
	})
}

func RejectFriendRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	fmt.Println(id)

	token := c.Cookies("jwt_token")

	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	usr := middlewares.VerifyTokenAndParse(token)

	if usr == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized",
		})
	}

	var friendRequest models.Friends

	// delete friend request

	affected := db.GetDB().Where("sender_id = ? AND receiver_id = ?", id, int(usr["id"].(float64))).First(&friendRequest)

	if affected.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Friend request not found",
		})
	}

	db.GetDB().Where("receiver_id = ? AND sender_id = ?", int(usr["id"].(float64)), id).Delete(&friendRequest)

	return c.Status(200).JSON(fiber.Map{
		"data": "Friend request deleted",
	})
}

func ServePhoto(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	t := middlewares.VerifyTokenAndParse(token)

	if t == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized, please log in",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", int(t["id"].(float64))).First(&user)

	if user.ID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	if user.ProfilePic == "" {
		return c.Status(404).JSON(fiber.Map{
			"data": "Photo not found",
		})
	}

	photoPath := user.ProfilePic

	return c.Status(201).SendFile(photoPath)

}

func GetUser(c *fiber.Ctx) error {

	var user models.User

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
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

	var friends []models.Friends

	db.GetDB().Where("(sender_id = ? OR receiver_id = ?) AND status = ?", user.ID, user.ID, "accepted").Find(&friends)

	var responseFriends []map[string]interface{}

	for _, friend := range friends {
		responseFriend := make(map[string]interface{})

		if friend.SenderID == int(user.ID) {
			responseFriend["id"] = friend.ReceiverID
			responseFriend["name"] = friend.ReceiverName
		} else {
			responseFriend["id"] = friend.SenderID
			responseFriend["name"] = friend.SenderName
		}

		responseFriends = append(responseFriends, responseFriend)
	}

	response["friendsNumber"] = len(responseFriends)
	response["friends"] = responseFriends

	var friendsPending []models.Friends

	db.GetDB().Where("receiver_id = ? AND status = ?", user.ID, "pending").Find(&friendsPending)

	var responseFriendsPending []map[string]interface{}

	for _, friendPending := range friendsPending {
		responseFriendPending := make(map[string]interface{})

		responseFriendPending["id"] = friendPending.SenderID
		responseFriendPending["name"] = friendPending.SenderName

		responseFriendsPending = append(responseFriendsPending, responseFriendPending)
	}

	response["friendsPendingNumber"] = len(responseFriendsPending)
	response["friendsPending"] = responseFriendsPending

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
		"data": response,
	})

}

func DeleteUser(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	db.GetDB().Delete(&user)

	return c.JSON(fiber.Map{
		"data": "User deleted successfully",
	})

}

func PromoteToLibrarian(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	user.Librarian = true

	db.GetDB().Save(&user)

	return c.JSON(fiber.Map{
		"data": fmt.Sprintf("User %s has been promoted to librarian", user.Name),
	})

}

func ModifyUser(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" {
		return c.Status(401).JSON(fiber.Map{
			"data": "User not found",
		})
	}

	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"data": "Error at parsing json",
		})
	}

	var user models.User

	db.GetDB().Where("id = ?", id).First(&user)

	if user.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User not found",
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
		"data": "User has been modified",
	})

}
