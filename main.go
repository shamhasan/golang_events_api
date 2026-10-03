package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"sham.com/app-server/config"
	"sham.com/app-server/controllers"
	"sham.com/app-server/middlewares"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()

	server := gin.Default()

	api := server.Group("/api")
	{
		api.GET("/events", controllers.GetEvents)
		api.GET("/events/:id", controllers.GetEventByID)

		api.POST("/auth/register", controllers.RegisterUser)
		api.POST("/auth/login", controllers.LoginUser)

		protected := api.Group("/")
		protected.Use(middlewares.RequiredAuth())
		{
			protected.GET("/events/user", controllers.GetEventsByUser)
			protected.GET("/auth/me", controllers.GetCurrentUser)
			protected.POST("/events", controllers.CreateEvent)
			protected.PUT("/events/:id", controllers.UpdateEvent)
			protected.DELETE("/events/:id", controllers.DeleteEvent)
		}
	}

	server.Run(":8080")
}
