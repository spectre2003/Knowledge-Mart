package controllers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	//"github.com/joho/godotenv"
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"knowledgeMart/utils"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var Validate *validator.Validate

func EmailSignup(c *gin.Context) {
	var Signup models.EmailSignupRequest

	if err := c.ShouldBindJSON(&Signup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process the incoming request" + err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(Signup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	if Signup.Password != Signup.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "passwords doesn't match",
		})
		return
	}

	otp := uint64(rand.Intn(900000) + 100000)
	otpExpiry := time.Now().Add(10 * time.Minute)

	User := models.User{
		Name:        Signup.Name,
		Email:       Signup.Email,
		PhoneNumber: Signup.PhoneNumber,
		Blocked:     false,
		Password:    Signup.Password,
		OTP:         otp,
		OTPExpiry:   otpExpiry,
		IsVerified:  false,
	}

	tx := database.DB.Where("email = ? AND deleted_at IS NULL", Signup.Email).First(&User)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to retreive information from the database",
		})
		return
	} else if tx.Error == gorm.ErrRecordNotFound {
		tx = database.DB.Create(&User)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to create a new user",
			})
			fmt.Println(tx.Error)
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user already exist",
		})
		return
	}
	err = sendOTPEmail(User.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Email login successful, please login to complete your email verification",
		"data": gin.H{
			"name":         User.Name,
			"email":        User.Email,
			"phone_number": User.PhoneNumber,
			"picture":      User.Picture,
			"address":      User.Address,
			"block_status": User.Blocked,
			"verified":     User.IsVerified,
		},
	})
	//c.Next()
}
func EmailLogin(c *gin.Context) {
	var LoginRequest models.EmailLoginRequest

	if err := c.ShouldBindJSON(&LoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(LoginRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	var User models.User
	tx := database.DB.Where("email = ? AND deleted_at is NULL", LoginRequest.Email).First(&User)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
		})
		return
	}

	if User.Password != LoginRequest.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Incorrect password",
		})
		return
	}

	token, err := utils.GenerateJWT(User.ID, User.Email, "user")
	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to generate token",
		})
		return
	}

	fmt.Println(token)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Login successful",
		"data": gin.H{
			"token":        token,
			"name":         User.Name,
			"email":        User.Email,
			"phone_number": User.PhoneNumber,
			"picture":      User.Picture,
			"address":      User.Address,
			"block_status": User.Blocked,
			"verified":     User.IsVerified,
		},
	})

}

func sendOTPEmail(to string, otp uint64) error {
	from := "knowledgemartv01@gmail.com"
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	appPassword := os.Getenv("SMTPAPP")
	fmt.Println(from)

	auth := smtp.PlainAuth("", from, appPassword, "smtp.gmail.com")

	msg := []byte("Subject: Verify your email\n\n" +
		fmt.Sprintf("Your OTP is %d", otp))
	err = smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
	fmt.Println("send OTP:", otp)
	if err != nil {
		fmt.Printf("Error in sending email: %v\n", err)
		return errors.New("failed to send email ")
	}
	return nil

}

func VarifyEmail(c *gin.Context) {

	email := c.Query("email")
	otpParam := c.Query("otp")

	otpParam = strings.TrimSpace(otpParam)

	fmt.Println("Received email:", email)
	fmt.Println("Received OTP:", otpParam)

	otp, err := strconv.Atoi(otpParam)
	if err != nil || email == "" || otp == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	var User models.User
	if err := database.DB.Where("email = ?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if User.OTP != uint64(otp) || time.Now().After(User.OTPExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid or expired OTP"})
		return
	}

	User.IsVerified = true
	if err := database.DB.Save(&User).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
}
