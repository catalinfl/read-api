package controllers

import (
	"strconv"

	"github.com/catalinfl/readit-api/db"
	"github.com/catalinfl/readit-api/models"
	"github.com/gofiber/fiber/v2"
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

	userBookMap := request.(map[string]interface{})

	userBook.UserID = uint(userBookMap["user_id"].(float64))
	userBook.BookID = uint(userBookMap["book_id"].(float64))
	userBook.Title = userBookMap["title"].(string)
	userBook.Author = userBookMap["author"].(string)
	userBook.Year = uint(userBookMap["year"].(float64))
	userBook.ISBN = userBookMap["isbn"].(string)
	userBook.Pages = uint(userBookMap["pages"].(float64))
	userBook.PagesRead = uint(userBookMap["pages_read"].(float64))
	userBook.Genre = userBookMap["genre"].(string)

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
		"message": "User book created successfully",
		"data":    userBook,
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

	db.GetDB().Save(&userBook)

	return c.JSON(fiber.Map{
		"message":    "User book updated successfully",
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

	var book models.UserBooks

	bookMap := request

	db.GetDB().Where("user_books_id = ?", bookMap["user_books_id"]).First(&book)

	if book.UserBooksID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"data": "Book doesn't exist",
		})
	}

	book.Genre = bookMap["genre"].(string)

	db.GetDB().Save(&book)

	return c.JSON(fiber.Map{
		"message": "Book genre updated successfully",
		"data":    book.Genre,
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
		"message": "User book deleted successfully",
	})

}
