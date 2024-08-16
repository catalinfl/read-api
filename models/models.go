package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MultiString []string

type Book struct {
	ID          int         `gorm:"primaryKey" json:"id"`
	Title       string      `gorm:"size:100" json:"title"`
	Author      string      `gorm:"size:100" json:"author"`
	Year        string      `gorm:"size:100" json:"year"`
	ISBN        string      `gorm:"size:100" json:"isbn"`
	Language    string      `gorm:"size:100" json:"language"`
	Pages       uint        `json:"pages"`
	Genre       string      `gorm:"size:100" json:"genre"`
	Publisher   string      `gorm:"size:100" json:"publisher"`
	Description string      `gorm:"size:1000" json:"description"`
	Photos      MultiString `gorm:"type:text" json:"photos"`
	Users       []User      `gorm:"many2many:user_books"`
}

func (ms MultiString) Value() (driver.Value, error) {
	return json.Marshal(ms)
}

func (ms *MultiString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)

	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ms)
}

func (MultiString) GormData() string {
	return "jsonb"
}

func (MultiString) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "text"
	case "postgres":
		return "jsonb"
	}
	return ""
}

type User struct {
	ID         int    `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"size:40" json:"name"`
	RealName   string `gorm:"size:60" json:"real_name"`
	Email      string `gorm:"size:40" json:"email"`
	Password   string `gorm:"size:100" json:"password"`
	Rank       string `gorm:"size:20" json:"rank"`
	Books      []Book `gorm:"many2many:user_books"`
	Librarian  bool   `json:"librarian"`
	Admin      bool   `json:"admin"`
	ProfilePic string `json:"profile_pic"`
}

type UserBooks struct {
	UserBooksID int    `gorm:"primaryKey" json:"user_books_id" db:"user_books.id"`
	UserID      uint   `json:"user_id" db:"user.id"`
	BookID      uint   `json:"book_id" db:"book.id"`
	Title       string `json:"title" db:"book.title"`
	Author      string `json:"author" db:"book.author"`
	Year        uint   `json:"year" db:"book.year"`
	ISBN        string `json:"isbn" db:"book.isbn"`
	Pages       uint   `json:"pages" db:"book.pages"`
	PagesRead   uint   `json:"pages_read" db:"pages_read"`
	Genre       string `json:"genre" db:"book.genre"`
}

func MigrateBooks(db *gorm.DB) {
	err := db.AutoMigrate(&Book{}, &User{}, &UserBooks{})

	if err != nil {
		panic(err)
	}

	fmt.Println("Books migration has been processed")
}
