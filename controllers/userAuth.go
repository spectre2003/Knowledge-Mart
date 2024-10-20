package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var Validate *validator.Validate

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/api/v1/googlecallback",
	ClientID:     os.Getenv("CLIENTID"),
	ClientSecret: os.Getenv("CLIENTSECRET"),
	Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint: google.Endpoint,
}

func GoogleHandleLogin(c *gin.Context) {
	state := "hjdfyuhadVFYU6781235"
	url := googleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
	c.Next()
}

func GoogleHandleCallback(c *gin.Context) {
	fmt.Println("Starting to handle callback")
	fmt.Printf("Callback URL Params: %v\n", c.Request.URL.Query())

	//code := c.Query("code")
	code := strings.TrimSpace(c.Query("code"))

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "missing code parameter",
		})
		return
	}
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token Exchange Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to exchange token",
		})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to get user information",
		})
		return
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to read user information",
		})
		return
	}

	var googleUser models.GoogleResponse
	err = json.Unmarshal(content, &googleUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to parse user information",
		})
		return
	}

	var existingUser models.User
	if err := database.DB.Where("email = ?", googleUser.Email).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			newUser := models.User{
				Email:       googleUser.Email,
				Name:        googleUser.Name,
				Picture:     googleUser.Picture,
				LoginMethod: "google",
				IsVerified:  true,
			}
			if err := database.DB.Create(&newUser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "failed",
					"message": "failed to create new user",
				})
				return
			}
			existingUser = newUser // Assign the newly created user to existingUser for later token generation
		} else {
			// Some other error occurred
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to fetch user from database",
			})
			return
		}
	}

	// Check if the user is blocked or needs to login with another method
	if existingUser.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user is unauthorized to access",
		})
		return
	}

	// Generate JWT using userID
	tokenstring, err := utils.GenerateJWT(existingUser.ID, "user")
	if tokenstring == "" || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "failed to create authorization token",
		})
		return
	}

	// Return success response with JWT and user info
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "login successful",
		"data": gin.H{
			"token": tokenstring,
			"user":  existingUser,
		},
	})
}

func EmailSignup(c *gin.Context) {

	var Signup models.EmailSignupRequest

	if err := c.ShouldBindJSON(&Signup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process the incoming request" + err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(Signup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	if Signup.Password != Signup.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "passwords doesn't match",
		})
		return
	}
	hashpassword, err := HashPassword(Signup.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "error in password hashing" + err.Error(),
		})
		return
	}

	refCode := utils.GenerateRandomString(5)

	otp, otpExpiry := GenerateOTP()

	User := models.User{
		Name:         Signup.Name,
		Email:        Signup.Email,
		PhoneNumber:  Signup.PhoneNumber,
		Blocked:      false,
		Password:     hashpassword,
		ReferralCode: refCode,
		OTP:          otp,
		OTPExpiry:    otpExpiry,
		IsVerified:   false,
		LoginMethod:  "email",
	}

	tx := database.DB.Where("email = ? AND deleted_at IS NULL", Signup.Email).First(&User)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retreive information from the database",
		})
		return
	} else if tx.Error == gorm.ErrRecordNotFound {
		tx = database.DB.Create(&User)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to create a new user",
			})
			fmt.Println(tx.Error)
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
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
		"status":  "success",
		"message": "Email login successful, please login to complete your email verification",
		"data": gin.H{
			"name":         User.Name,
			"email":        User.Email,
			"phone_number": User.PhoneNumber,
			"picture":      User.Picture,
			"block_status": User.Blocked,
			"verified":     User.IsVerified,
		},
	})
}
func EmailLogin(c *gin.Context) {
	var LoginRequest models.EmailLoginRequest

	if err := c.ShouldBindJSON(&LoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	Validate = validator.New()

	err := Validate.Struct(LoginRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var User models.User
	tx := database.DB.Where("email = ? AND deleted_at is NULL", LoginRequest.Email).First(&User)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid email or password",
		})
		return
	}

	err = CheckPassword(User.Password, LoginRequest.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "Incorrect password",
		})
		return
	}

	if User.Blocked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user is not authorized to access",
		})
		return
	}

	if User.ReferralCode == "" {
		refCode := utils.GenerateRandomString(5)
		User.ReferralCode = refCode

		if err := database.DB.Model(&User).Update("referral_code", User.ReferralCode).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "failed to update referral code",
			})
			return
		}
	}

	token, err := utils.GenerateJWT(User.ID, "user")
	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to generate token",
		})
		return
	}

	fmt.Println(token)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"token":        token,
			"id":           User.ID,
			"name":         User.Name,
			"email":        User.Email,
			"phone_number": User.PhoneNumber,
			"picture":      User.Picture,
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
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var User models.User
	if err := database.DB.Where("email = ?", email).First(&User).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "User not found",
		})
		return
	}

	if User.OTP != uint64(otp) || time.Now().After(User.OTPExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid or expired OTP",
		})
		return
	}

	User.IsVerified = true
	if err := database.DB.Save(&User).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to update user verification status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Email verified",
	})
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
			"status":  "failed",
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
				"status":  "failed",
				"message": "user not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
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
			"status":  "failed",
			"message": "failed to update OTP in the database",
		})
		return
	}

	err := sendOTPEmail(user.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to send OTP email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "OTP has been resent successfully",
	})

}
