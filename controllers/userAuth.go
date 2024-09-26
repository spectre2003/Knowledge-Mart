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
	"golang.org/x/crypto/bcrypt"
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
	hashpassword, err := HashPassword(Signup.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "error in password hashing" + err.Error(),
		})
		return
	}

	otp, otpExpiry := GenerateOTP()

	User := models.User{
		Name:        Signup.Name,
		Email:       Signup.Email,
		PhoneNumber: Signup.PhoneNumber,
		Blocked:     false,
		Password:    hashpassword,
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

	err = CheckPassword(User.Password, LoginRequest.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Incorrect password",
		})
		return
	}

	// if User.Password != LoginRequest.Password {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"status":  false,
	// 		"message": "Incorrect password",
	// 	})
	// 	return
	// }
	if User.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "user is not authorized to access",
		})
		return
	}

	token, err := utils.GenerateJWT(User.ID, "user")
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

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateOTP() (uint64, time.Time) {
	otp := uint64(rand.Intn(900000) + 100000)
	otpExpiry := time.Now().Add(3 * time.Minute)
	return otp, otpExpiry
}

func ResendOTP(c *gin.Context) {

	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "email is required",
		})
		return
	}

	otp, otpExpiry := GenerateOTP()

	var user models.User
	tx := database.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "user not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "error retrieving user from the database",
			})
		}
		return
	}

	user.OTP = otp
	user.OTPExpiry = otpExpiry

	tx = database.DB.Save(&user)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update OTP in the database",
		})
		return
	}

	err := sendOTPEmail(user.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to send OTP email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "OTP has been resent successfully",
	})

}
