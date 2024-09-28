package controllers

import (
	database "knowledgeMart/config"
	"knowledgeMart/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func GetUserProfile(c *gin.Context) {
	var user models.User

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	userIDStr, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	if err := database.DB.Where("id = ? AND deleted_at IS NULL", userIDStr).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	userProfile := models.UserProfileResponse{
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Picture:     user.Picture,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully retrieved user profile",
		"data": gin.H{
			"profile": userProfile,
		},
	})
}

func EditUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var Request models.EditUserProfileResponse

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}
	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var existingUser models.User

	if err := database.DB.Where("id = ?", Request.ID).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch user details from the database",
		})
		return
	}

	existingUser.Name = Request.Name

	if existingUser.Email != Request.Email {
		otp, otpExpiry := GenerateOTP()
		existingUser.OTP = otp
		existingUser.OTPExpiry = otpExpiry

		existingUser.Email = Request.Email
		existingUser.IsVerified = false

		tx := database.DB.Model(&existingUser).Updates(models.User{
			OTP:        existingUser.OTP,
			OTPExpiry:  existingUser.OTPExpiry,
			Email:      existingUser.Email,
			IsVerified: existingUser.IsVerified,
		})
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to change the block status ",
			})
			return
		}

		err := sendOTPEmail(existingUser.Email, otp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "failed",
				"message": "failed to send verification email",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Email has been changed, please verify the new email to complete the update",
		})
		return
	}

	existingUser.PhoneNumber = Request.PhoneNumber

	if err := database.DB.Model(&existingUser).Select("name", "email", "phone_number").Updates(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update user profile",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated user information",
		"data": gin.H{
			"product": Request,
		},
	})

}

func AddAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var request models.AddAddresRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
	}
	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	var user models.User
	if err := database.DB.Where("id = ?", userIDUint).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var UserAddresses []models.Address
	if err := database.DB.Where("user_id = ?", userIDUint).Find(&UserAddresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve the existing user addresses from the database",
		})
		return
	}

	//check if the user already has 3 address...3 is the limit
	if len(UserAddresses) >= 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "user already have three addresses, please delete or edit the existing addresses",
		})
		return
	}

	newAddress := models.Address{
		UserID:       userIDUint,
		StreetName:   request.StreetName,
		StreetNumber: request.StreetNumber,
		City:         request.City,
		State:        request.State,
		PinCode:      request.Pincode,
	}
	if err := database.DB.Create(&newAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "false",
			"message": "failed to create address" + err.Error(),
		})
		return
	}
	addressResponse := models.AddressResponse{
		Username:     user.Name,
		StreetName:   newAddress.StreetName,
		StreetNumber: newAddress.StreetNumber,
		City:         newAddress.City,
		State:        newAddress.State,
		PinCode:      newAddress.PinCode,
		Phone:        user.PhoneNumber,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully added new address",
		"data": gin.H{
			"id":      newAddress.ID,
			"address": addressResponse,
		},
	})
}

func ListAllAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	var user models.User
	if err := database.DB.Where("id = ?", userIDUint).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}
	var Addresses []models.Address
	var AddressResponse []models.AddressResponse

	tx := database.DB.Select("*").Find(&Addresses)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "failed to retrieve data from the database",
		})
		return
	}

	for _, address := range Addresses {
		AddressResponse = append(AddressResponse, models.AddressResponse{
			ID:           address.ID,
			Username:     user.Name,
			StreetName:   address.StreetName,
			StreetNumber: address.StreetNumber,
			City:         address.City,
			State:        address.State,
			PinCode:      address.PinCode,
			Phone:        user.PhoneNumber,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully get all the addresses",
		"data": gin.H{
			"address": AddressResponse,
		},
	})
}

func EditAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	var Request models.EditAddresRequest

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to process request",
		})
		return
	}
	validate := validator.New()
	if err := validate.Struct(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	var existingAddress models.Address

	if err := database.DB.Where("id = ?", Request.ID).First(&existingAddress).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "failed to fetch address from the database",
		})
		return
	}

	existingAddress.StreetName = Request.StreetName
	existingAddress.StreetNumber = Request.StreetNumber
	existingAddress.City = Request.City
	existingAddress.State = Request.State
	existingAddress.PinCode = Request.Pincode

	if err := database.DB.Model(&existingAddress).Select("street_name", "street_number", "city", "state", "pincode").Updates(&existingAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to update address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated address information",
		"data": gin.H{
			"product": Request,
		},
	})

}

func DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "failed",
			"message": "user not authorized ",
		})
		return
	}

	_, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "failed to retrieve user information",
		})
		return
	}

	addressIDStr := c.Query("addressid")
	addressID, err := strconv.Atoi(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "invalid address ID",
		})
		return
	}

	var address models.Address

	if err := database.DB.First(&address, addressID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "failed",
			"message": "address is not present in the database",
		})
		return
	}

	if err := database.DB.Delete(&address, addressID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "unable to delete the address from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully deleted the address",
	})

}