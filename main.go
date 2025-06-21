package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"html/template"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"example.com/hello-gin/db"
	"example.com/hello-gin/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Initialize the Haiku store with the mongoClient
	db.InitHaikuStore(mongoClient)
	log.Println("Haiku store initialized.")

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

	// Haiku routes
	r.POST("/haikus", handleAddHaiku)
	r.GET("/haikus", handleGetHaikusHTML)

	log.Println("Starting Gin server on 0.0.0.0:8080...")
	if err := r.Run(); err != nil {
		log.Fatalf("FATAL: Gin server failed to start: %v", err)
	}
}

// handleAddHaiku handles POST requests to add a new haiku.
func handleAddHaiku(c *gin.Context) {
	var haikuRequest models.Haiku
	if err := c.ShouldBindJSON(&haikuRequest); err != nil {
		log.Printf("Error binding JSON for new haiku: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Set required fields if not provided or to ensure correctness
	haikuRequest.ID = primitive.NewObjectID() // Generate new ID
	if haikuRequest.Date.IsZero() {
		haikuRequest.Date = time.Now() // Default to current time if not provided
	}

	if err := db.AddHaiku(haikuRequest); err != nil {
		log.Printf("Error adding haiku to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add haiku to database"})
		return
	}

	log.Printf("Successfully added haiku via API: %s", haikuRequest.Text)
	c.JSON(http.StatusCreated, haikuRequest)
}

// handleGetHaikusHTML handles GET requests to display haikus in an HTML page.
func handleGetHaikusHTML(c *gin.Context) {
	haikus, err := db.GetAllHaikus()
	if err != nil {
		log.Printf("Error getting all haikus from database: %v", err)
		c.HTML(http.StatusInternalServerError, "", "Error retrieving haikus: "+err.Error()) // Basic error
		return
	}

	// Simple HTML template embedded as a string
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Haikus</title>
    <style>
        body { font-family: sans-serif; margin: 20px; background-color: #f4f4f4; color: #333; }
        h1 { color: #4a4a4a; text-align: center; }
        .haiku-container { background-color: #fff; border: 1px solid #ddd; border-radius: 5px; padding: 15px; margin-bottom: 15px; box-shadow: 2px 2px 5px rgba(0,0,0,0.1); }
        .haiku-text { font-style: italic; font-size: 1.1em; margin-bottom: 10px; white-space: pre-wrap; }
        .haiku-meta { font-size: 0.9em; color: #555; }
        .haiku-meta p { margin: 5px 0; }
        .no-haikus { text-align: center; font-size: 1.2em; color: #777; margin-top: 30px; }
    </style>
</head>
<body>
    <h1>Haiku Collection</h1>
    {{if .}}
        {{range .}}
        <div class="haiku-container">
            <div class="haiku-text">{{.Text}}</div>
            <div class="haiku-meta">
                <p><strong>Date:</strong> {{.Date.Format "2006-01-02 15:04:05"}}</p>
                <p><strong>Moisture:</strong> {{printf "%d" .Moisture}} | <strong>Temperature:</strong> {{printf "%d°C" .Temperature}} | <strong>Illumination:</strong> {{printf "%d lux" .Illumination}} | <strong>pH:</strong> {{printf "%d" .PH}}</p>
            </div>
        </div>
        {{end}}
    {{else}}
        <p class="no-haikus">No haikus found. Try adding some!</p>
    {{end}}
</body>
</html>`

	// Using text/template for simplicity, html/template is safer for user-generated content
	// but for this controlled output, text/template is fine.
	// For robustness, consider html/template.New("haikus").Parse(htmlTemplate)
	tmpl, err := template.New("haikusPage").Parse(htmlTemplate)
	if err != nil {
		log.Printf("Error parsing HTML template: %v", err)
		c.HTML(http.StatusInternalServerError, "", "Error rendering page")
		return
	}

	c.Status(http.StatusOK) // Set status header first
	// Set content type explicitly, though Gin might do it.
	// For HTML served directly by `c.HTML`, Gin sets Content-Type to text/html.
	// If using ExecuteTemplate directly to c.Writer, ensure Content-Type is set.
	// c.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(c.Writer, haikus)
	if err != nil {
		log.Printf("Error executing HTML template: %v", err)
		// Don't try to write another HTML error if headers are already sent.
		// c.HTML(http.StatusInternalServerError, "", "Error rendering page content")
		return
	}
	log.Printf("Successfully served HTML page with %d haikus.", len(haikus))
}
