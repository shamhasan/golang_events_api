package controllers

import (
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"gorm.io/gorm"
	"sham.com/app-server/config"
	"sham.com/app-server/models"
)

func InitImageKit() *imagekit.Client {
	client := imagekit.NewClient(
		option.WithPrivateKey(os.Getenv("IMAGEKIT_PRIVATE_KEY")),
	)
	return &client
}

func CreateEvent(c *gin.Context) {
	userId, _ := c.Get("userId")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Gambar wajib diupload",
			"error":   err.Error()})
		return
	}
	defer file.Close()

	fileName := header.Filename
	ik := InitImageKit()
	uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File:     file,
		FileName: fileName,
	})
	if errUpload != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengupload gambar imageKit",
		})
		return
	}

	parsedTime, _ := time.Parse(time.RFC3339, c.PostForm("dateTime"))

	event := models.Event{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Location:    c.PostForm("location"),
		Image:       uploadRes.URL,
		DateTime:    parsedTime,
		ImageId:     uploadRes.FileID,
		UserId:      userId.(int),
	}

	config.DB.Create(&event)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"event":   event,
	})
}

func GetEvents(c *gin.Context) {
	var events []models.Event

	// inisiasi query
	query := config.DB.Model(&models.Event{})

	// filter by query parameters
	search := c.Query("search")

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// hitung total sebelum limit pagination
	var totalRows int64
	query.Count(&totalRows)

	// tangkap parameter query
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "6")

	// konversi parameter query ke integer
	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page < 1 {
		page = 1
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil || page < 1 {
		limit = 6
	}

	// hitung offset

	offset := (page - 1) * limit

	// hitung halaman
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))

	// eksekusi query dengan limit dan offset
	if errEvents := query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Limit(limit).Offset(offset).Find(&events).Error; errEvents != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event retrieved successfully",
		"events":  events,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"totalRows":  totalRows,
			"totalPages": totalPages,
		},
	})
}

func GetEventByID(context *gin.Context) {
	var event models.Event
	paramsId := context.Param("id")

	var eventData = config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).First(&event, paramsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"message": "Event not found",
			"error":   eventData.Error(),
		})

		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Detail Event retrieved successfully",
		"events":  event,
	})
}

func GetEventsByUser(c *gin.Context) {
	var events []models.Event
	userId, _ := c.Get("userId")

	errEvents := config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id", userId).Find(&events).Error
	if errEvents != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Error tidak ditemuka" + errEvents.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":  events,
		"message": "Event retrieved successfully",
	})
}

func UpdateEvent(c *gin.Context) {
	userId, _ := c.Get("userId")
	var event models.Event

	paramsId := c.Param("id")
	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Event not found",
			"error":   eventData.Error(),
		})

		return
	}

	if event.UserId != userId.(int) {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "You are not authorized to update this event",
		})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()

		ik := InitImageKit()
		// upload new image
		fileName := header.Filename
		uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
			File:     file,
			FileName: fileName,
		})
		if errUpload == nil {
			// delete old image
			if event.ImageId != "" {
				ik.Files.Delete(context.Background(), event.ImageId)
			}
			event.Image = uploadRes.URL
			event.ImageId = uploadRes.FileID
		}
	}

	if name := c.PostForm("name"); name != "" {
		event.Name = name
	}
	if description := c.PostForm("description"); description != "" {
		event.Description = description
	}
	if location := c.PostForm("location"); location != "" {
		event.Location = location
	}
	if dateTimeStr := c.PostForm("dateTime"); dateTimeStr != "" {
		parseTime, errParse := time.Parse(time.RFC3339, dateTimeStr)
		if errParse == nil {
			event.DateTime = parseTime
		}
	}

	config.DB.Save(&event)
	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"event":   event,
	})
}

func DeleteEvent(c *gin.Context) {
	userId, _ := c.Get("userId")
	var event models.Event

	paramsId := c.Param("id")
	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Event not found",
			"error":   eventData.Error(),
		})

		return
	}
	if event.UserId != userId.(int) {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "You are not authorized to update this event",
		})

		return
	}

	if event.ImageId != "" {
		ik := InitImageKit()
		ik.Files.Delete(context.Background(), event.ImageId)
	}

	config.DB.Unscoped().Delete(&event)
	c.JSON(http.StatusOK, gin.H{
		"message": "Event deleted successfully",
	})
}
