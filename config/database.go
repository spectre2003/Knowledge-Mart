package database

import (
	"fmt"
	"knowledgeMart/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dsn := fmt.Sprintf("host=127.0.0.1 user=postgres password=password dbname=knowledgemart port=5432 sslmode=disable")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	} else {
		fmt.Println("connection to database :OK")
	}
	DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Seller{},
		&models.Product{},
		&models.Category{},
		&models.Address{},
	)
}
