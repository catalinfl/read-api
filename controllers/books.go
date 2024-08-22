package controllers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/middlewares"
	"github.com/catalinfl/readit-api/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetBooks(c *fiber.Ctx) error {

	getInfiniteScrollBooks(c)

	return nil
}

func GetBooksPaginated(c *fiber.Ctx) error {

	getBooksPagination(c)

	return nil
}

func CreateBook(c *fiber.Ctx) error {

	var book models.Book

	if err := c.BodyParser(&book); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	if len(book.Title) < 3 && len(book.Title) > 100 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Title must be between 3 and 100 characters",
		})
	}

	var existingBook models.Book

	db.GetDB().Where("title = ?", book.Title).First(&existingBook)

	if existingBook.ID > 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Book already exists",
		})
	}

	db.GetDB().Create(&book)

	return c.JSON(fiber.Map{
		"data": "Book created successfully",
	})
}

func GetAllBooks(c *fiber.Ctx) error {

	var books []models.Book

	db.GetDB().Find(&books)

	return c.JSON(fiber.Map{
		"data": books,
	})
}

func getInfiniteScrollBooks(c *fiber.Ctx) error {

	bookId := c.Params("id")

	currentId, err := strconv.Atoi(bookId)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid ID",
		})
	}

	var infiniteScrollBooks []models.Book

	db.GetDB().Where("id >= ?", currentId).Limit(5).Find(&infiniteScrollBooks)

	return c.JSON(fiber.Map{
		"data":       infiniteScrollBooks,
		"nextCursor": currentId + 5,
	})
}

func getBooksPagination(c *fiber.Ctx) {

	page, _ := strconv.Atoi(c.Query("page", "1"))

	pageSize := 5

	if page < 1 {
		page = 1
	}

	var books []models.Book

	var totalBooks int64

	db.GetDB().Model(&models.Book{}).Count(&totalBooks)

	db.GetDB().Offset((page - 1) * pageSize).Limit(pageSize).Find(&books)

	hasMore := int(totalBooks) > page*pageSize

	c.Status(200).JSON(
		fiber.Map{
			"data":    books,
			"hasMore": hasMore,
		},
	)

}

func GetUserBooks(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var userBooks []models.UserBooks

	db.GetDB().Where("user_id = ?", id).Find(&userBooks)

	return c.JSON(fiber.Map{
		"data": userBooks,
	})
}

func GetAllUserBooks(c *fiber.Ctx) error {
	var userBooks []models.UserBooks

	db.GetDB().Find(&userBooks)

	return c.JSON(fiber.Map{
		"data": userBooks,
	})
}

func CreateUserBook(c *fiber.Ctx) error {
	var request interface{}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var userBook models.UserBooks

	token := c.Cookies("jwt_token")

	t := middlewares.VerifyTokenAndParse(token)

	if t == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized, please log in",
		})
	}

	fmt.Println(t)

	userBookMap := request.(map[string]interface{})

	var user models.User

	db.GetDB().Where("name = ?", t).First(&user)

	userBook.UserID = uint(user.ID)
	userBook.BookID = uint(userBookMap["book_id"].(float64))
	userBook.PagesRead = uint(userBookMap["pages_read"].(float64))

	var existingUserBook models.UserBooks

	var existingBook models.Book

	db.GetDB().Where("id = ?", userBook.BookID).First(&existingBook)

	if existingBook.ID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Book doesn't exist",
		})
	}

	db.GetDB().Where("user_id = ? AND book_id = ?", userBook.UserID, userBook.BookID).First(&existingUserBook)

	if existingUserBook.UserBooksID > 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "You already have this book",
		})
	}

	db.GetDB().Create(&userBook)

	return c.JSON(fiber.Map{
		"data": "User book created successfully",
	})

}

func DeleteBook(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	db.GetDB().Delete(&book)

	return c.JSON(fiber.Map{
		"data": "Book deleted successfully",
	})
}

func UpdateReadingBook(c *fiber.Ctx) error {

	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var userBook models.UserBooks

	userBookMap := request

	db.GetDB().Where("user_books_id = ?", userBookMap["user_books_id"]).First(&userBook)

	userBook.PagesRead = uint(userBookMap["pages_read"].(float64))

	db.GetDB().Model(&userBook).Updates(userBook)

	return c.JSON(fiber.Map{
		"data":       "User book updated successfully",
		"pages_read": userBook.PagesRead,
	})

}

func UpdateGenre(c *fiber.Ctx) error {

	request := make(map[string]interface{})

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	bookMap := request

	db.GetDB().Where("id = ?", bookMap["id"]).First(&book)

	if book.ID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Book doesn't exist",
		})
	}

	book.Genre = bookMap["genre"].(string)

	db.GetDB().Save(&book)

	return c.JSON(fiber.Map{
		"data": book.Genre,
	})
}

func DeleteUserBook(c *fiber.Ctx) error {

	id := c.Params("bookId")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var userBook models.UserBooks

	db.GetDB().Where("user_books_id = ?", id).First(&userBook)

	if userBook.UserBooksID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "User book not found",
		})
	}

	db.GetDB().Delete(&userBook)

	return c.JSON(fiber.Map{
		"data": "User book deleted successfully",
	})

}

func ModifyBook(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	var request map[string]interface{}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	if err := db.GetDB().First(&book, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Book not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch book",
		})
	}

	for key, value := range request {
		switch key {
		case "title":
			book.Title = value.(string)
		case "author":
			book.Author = value.(string)
		case "year":
			book.Year = value.(string)
		case "isbn":
			book.ISBN = value.(string)
		case "language":
			book.Language = value.(string)
		case "pages":
			book.Pages = uint(value.(float64)) // JSON numbers are float64
		case "genre":
			book.Genre = value.(string)
		case "publisher":
			book.Publisher = value.(string)
		case "description":
			book.Description = value.(string)
		case "photos":
			book.Photos = value.([]string)
		}
	}

	if err := db.GetDB().Save(&book).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update book",
		})
	}

	return c.JSON(fiber.Map{
		"data": "Book updated successfully",
		"book": book,
	})
}

func DeleteBookLibrarian(c *fiber.Ctx) error {

	var book models.Book

	id := c.Params("bookId")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	db.GetDB().Delete(&book)

	return c.JSON(fiber.Map{
		"data": "Book deleted successfully",
	})

}

func AddPhotosForBooks(c *fiber.Ctx) error {

	token := c.Cookies("jwt_token")

	t := middlewares.VerifyTokenAndParse(token)

	if t == nil {
		return c.Status(401).JSON(fiber.Map{
			"data": "Unauthorized, please log in",
		})
	}

	bookId := c.Params("bookId")

	if bookId == "" || bookId == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", bookId).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	form, err := c.MultipartForm()

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	file := form.File["photos"]

	if len(file) != 1 {
		return c.Status(400).JSON(fiber.Map{
			"data": "You need to send just one photo",
		})
	}

	photo := file[0]

	if photo.Size > 1<<20 {
		return c.Status(400).JSON(fiber.Map{
			"data": "File size too big",
		})
	}

	dirPath := "/app/books/"

	filePath := dirPath + strconv.Itoa(int(book.ID))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {

		os.Mkdir(dirPath, os.ModePerm)

	}

	if err := c.SaveFile(photo, filePath); err != nil {
		fmt.Println(err)
		return c.Status(500).JSON(fiber.Map{
			"data": "Failed to save photo",
		})
	}

	emptyArray := []string{}

	// for future add more photos

	book.Photos = append(emptyArray, filePath)

	db.GetDB().Save(&book)

	return c.JSON(fiber.Map{
		"data": "Photo added successfully",
	})
}

func GetBooksPhoto(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	return c.Status(200).SendFile(book.Photos[0])

}

func GetBook(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": book,
	})

}

func DeleteBookPhoto(c *fiber.Ctx) error {

	id := c.Params("bookId")

	if id == "" || id == "0" {
		return c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	var book models.Book

	db.GetDB().Where("id = ?", id).First(&book)

	if book.ID == 0 {
		return c.Status(404).JSON(fiber.Map{
			"data": "Book not found",
		})
	}

	err := os.Remove(book.Photos[0])

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data": "Failed to delete photo",
		})
	}

	book.Photos = nil

	db.GetDB().Save(&book)

	return c.JSON(fiber.Map{
		"data": "Photo deleted successfully",
	})

}
