package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"example.com/hello-gin/db"
)

var mongoClient *mongo.Client

func main() {
	log.Println("Application starting...")
	var err error
	log.Println("Attempting to connect to database...")
	mongoClient, err = db.ConnectDB()
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to MongoDB after setup in db.ConnectDB(): %v", err)
	}
	log.Println("Database connection successful.")

	// Disconnect from MongoDB when the application closes
	// This part is a bit tricky with Gin's default Run() as it blocks.
	// For a production app, you'd handle graceful shutdown.
	// defer func() {
	// 	if err = mongoClient.Disconnect(context.TODO()); err != nil {
	// 		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	// 	}
	// }()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/db-ping", func(c *gin.Context) {
		if mongoClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "MongoDB client is not initialized",
			})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := mongoClient.Ping(ctx, readpref.Primary())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to ping MongoDB: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully pinged MongoDB!",
		})
	})

	log.Println("Starting Gin server on 0.0.0.0:8080...")
	if err := r.Run(); err != nil {
		log.Fatalf("FATAL: Gin server failed to start: %v", err)
	}
}
