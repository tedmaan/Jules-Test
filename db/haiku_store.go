package db

import (
	"context"
	"log"
	"os"
	"time"

	"example.com/hello-gin/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var haikuCollection *mongo.Collection

// InitHaikuStore initializes the database connection and collection for haikus.
// It's a simplified way to ensure the client is available.
// In a larger application, client management would be more robust.
func InitHaikuStore(mongoClient *mongo.Client) {
	client = mongoClient
	dbName := os.Getenv("MONGODB_DATABASE_NAME")
	if dbName == "" {
		dbName = "mydatabase" // Default database name if not set in .env
		log.Printf("Warning: MONGODB_DATABASE_NAME not set in .env, using default '%s'", dbName)
	}
	haikuCollection = client.Database(dbName).Collection("haikus")
	log.Printf("Haiku store initialized. Using database '%s' and collection 'haikus'.", dbName)
}

// AddHaiku inserts a new haiku into the database.
func AddHaiku(haiku models.Haiku) error {
	if haikuCollection == nil {
		log.Fatal("FATAL: haikuCollection is not initialized. Call InitHaikuStore first.")
		// In a real app, return an error instead of Fatal
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := haikuCollection.InsertOne(ctx, haiku)
	if err != nil {
		log.Printf("Error inserting haiku: %v", err)
		return err
	}
	log.Printf("Successfully inserted haiku with text: %s", haiku.Text)
	return nil
}

// GetAllHaikus retrieves all haikus from the database.
func GetAllHaikus() ([]models.Haiku, error) {
	if haikuCollection == nil {
		log.Fatal("FATAL: haikuCollection is not initialized. Call InitHaikuStore first.")
		// In a real app, return an error instead of Fatal
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var haikus []models.Haiku
	cursor, err := haikuCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "date", Value: -1}}))
	if err != nil {
		log.Printf("Error finding haikus: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &haikus); err != nil {
		log.Printf("Error decoding haikus: %v", err)
		return nil, err
	}

	if haikus == nil {
		log.Println("No haikus found, returning empty slice.")
		return []models.Haiku{}, nil // Return empty slice if no documents found
	}

	log.Printf("Successfully retrieved %d haikus.", len(haikus))
	return haikus, nil
}
