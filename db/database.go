package db

import (
	"fmt"
	"os"

	"github.com/catalinfl/readit-api/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var database *gorm.DB

func Connect() {
	godotenv.Load()

	data_link := os.Getenv("DB_POSTGRESQL_KEY")

	db, err := gorm.Open(postgres.Open(data_link), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	db.Exec("UPDATE user_books SET year = 2024 WHERE year = ''")

	models.MigrateBooks(db)

	database = db

	fmt.Println("Database is connected right now")
}

func GetDB() *gorm.DB {
	return database
}
