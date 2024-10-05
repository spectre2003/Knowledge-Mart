package database

import (
	"fmt"
	"knowledgeMart/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dsn := fmt.Sprintf("host=127.0.0.1 user=postgres password=password dbname=knowledgemart port=5432 sslmode=disable")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect to database")
	} else {
		fmt.Println("connection to database :OK")
	}
	err = DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Seller{},
		&models.Product{},
		&models.Category{},
		&models.Address{},
		&models.Cart{},
		&models.SellerRating{},
		&models.Order{},
	)
	if err != nil {
		fmt.Println("Migration failed:", err)
	}
	err = DB.AutoMigrate(
		&models.OrderItem{},
	)
	if err != nil {
		fmt.Println("Migration failed for OrderItem:", err)
	}

}
