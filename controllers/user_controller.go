package controllers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sham.com/app-server/config"
	"sham.com/app-server/models"
)

type AuthInputRegister struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthInputLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func RegisterUser(context *gin.Context) {
	var input AuthInputRegister

	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body, ada error pada input data",
			"error":   err.Error(),
		})
		return
	}

	// Check if email already exists in DB
	var existingUser models.User
	userData := config.DB.Where("email = ?", input.Email).First(&existingUser).Error
	if userData == nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "Email sudah terdaftar, silahkan gunakan email lain",
		})
		return
	}

	// Hashing password
	hashedPass, errHash := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if errHash != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal enkripsi password",
		})
		return
	}

	// Insert ke DB
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPass),
	}

	userCreated := config.DB.Create(&user).Error

	if userCreated != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "Email Mungkin sudah terdaftar, silahkan gunakan email lain",
			"error":   userCreated.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "events": user.Events},
	})
}

func LoginUser(context *gin.Context) {
	var input AuthInputLogin
	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body, ada error pada input data",
			"error":   err.Error(),
		})
		return
	}

	// Check if email exists in DB
	var user models.User
	userData := config.DB.Where("email = ?", input.Email).First(&user).Error

	if userData != nil {
		context.JSON(http.StatusUnauthorized, gin.H{
			"message": "Email tidak ditemukan",
			"error":   userData.Error(),
		})
		return
	}

	// Check if password matches
	errMatchPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))

	if errMatchPassword != nil {
		context.JSON(http.StatusUnauthorized, gin.H{
			"message": "Password salah",
			"error":   errMatchPassword.Error(),
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	fmt.Println("Secret saat login", os.Getenv("JWT_SECRET"))
	tokenString, errToken := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if errToken != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal membuat token",
			"error":   errToken.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   tokenString,
		"user":    gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "events": user.Events},
	})

}

func GetCurrentUser(context *gin.Context) {
	// Get userId from context
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
	}

	var user models.User
	userData := config.DB.Select("id", "name", "email").First(&user, userId).Error

	if userData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "User Not Found!",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"user": gin.H{"name": user.Name, "email": user.Email, "id": user.ID, "events": user.Events},
	})
}
