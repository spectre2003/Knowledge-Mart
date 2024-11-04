package database

import (
	"fmt"
	"knowledgeMart/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	fmt.Println("Connection to database: OK")

	err = DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Course{},
		&models.Seller{},
		&models.Product{},
		&models.Category{},
		&models.Address{},
		&models.Cart{},
		&models.Semester{},
		&models.SellerRating{},
		&models.Subject{},
		&models.WhishList{},
		&models.Payment{},
		&models.UserWallet{},
		&models.SellerWallet{},
		&models.CouponInventory{},
		&models.CouponUsage{},
		&models.UserReferralHistory{},
		&models.Note{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		fmt.Println("Migration failed:", err)
	} else {
		fmt.Println("Migrations: OK")
	}
}
