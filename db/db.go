package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ConnectDB loads environment variables and connects to MongoDB
func ConnectDB() (*mongo.Client, error) {
	log.Println("Attempting to load .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found or error loading it. Will rely on existing environment variables. Error:", err)
	} else {
		log.Println(".env file loaded successfully (if present).")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("FATAL: 'MONGODB_URI' environmental variable is not set. Please set it in .env or your environment. See README.md for more info.")
	}
	// Avoid logging the full URI with password in production, but for this debugging it's fine.
	log.Println("MONGODB_URI found. Attempting to connect...")


	log.Println("Creating new MongoDB client...")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Error creating MongoDB client: %v\n", err)
		return nil, err
	}

	log.Println("Connecting to MongoDB (timeout 10s)...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v\n", err)
		return nil, err
	}
	log.Println("MongoDB client connected.")

	log.Println("Pinging MongoDB primary node (timeout 10s)...")
	// Use a new context for ping, as the previous one might be near its deadline
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		log.Printf("Error pinging MongoDB: %v\n", err)
		// Optionally, you might want to disconnect the client here if ping fails
		// client.Disconnect(context.TODO())
		return nil, err
	}
	log.Println("Successfully connected and pinged MongoDB.")

	return client, nil
}
