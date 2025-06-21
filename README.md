# Go Gin MongoDB Example

This is a sample Go application using the Gin framework that connects to a MongoDB database.

## Prerequisites

- Go (version 1.21 or higher recommended)
- Docker and Docker Compose (for local MongoDB setup)
- Access to a MongoDB instance (either local via Docker or cloud-based like MongoDB Atlas)

## Setup

1.  **Clone the repository (if applicable):**
    ```bash
    git clone <repository-url>
    cd <repository-name>
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```
    Or, if you prefer to install them explicitly:
    ```bash
    go get github.com/gin-gonic/gin
    go get github.com/joho/godotenv
    go get go.mongodb.org/mongo-driver/mongo
    ```

3.  **Create a `.env` file:**
    Create a file named `.env` in the root of the project directory. Add your MongoDB connection string to this file:
    ```env
    MONGODB_URI="your_mongodb_connection_string_here"
    ```
    Replace `"your_mongodb_connection_string_here"` with your actual MongoDB URI.

    **For Local Development (using Docker):**
    The `docker-compose.yml` file included in this project will set up a local MongoDB container.
    The default URI for this local instance is:
    `MONGODB_URI="mongodb://localhost:27017/mydatabase"`

    If you modify `docker-compose.yml` to include a username and password (e.g., `MONGO_INITDB_ROOT_USERNAME: myuser`, `MONGO_INITDB_ROOT_PASSWORD: mypassword`), your URI would look like:
    `MONGODB_URI="mongodb://myuser:mypassword@localhost:27017/mydatabase?authSource=admin"`

    **For Cloud MongoDB (e.g., MongoDB Atlas):**
    `MONGODB_URI="mongodb+srv://<username>:<password>@<cluster-url>/<database_name>?retryWrites=true&w=majority"`

## Running the Application

1.  **(Optional) Start Local MongoDB Container:**
    If you're using the local Docker setup for MongoDB, navigate to the project root and run:
    ```bash
    docker-compose up -d
    ```
    This will start a MongoDB container in detached mode. The `-d` flag runs it in the background.
    To stop it later: `docker-compose down`

2.  **Start the Go application:**
    ```bash
    go run main.go
    ```
    The application will start, and you should see a log message indicating a successful connection to MongoDB. By default, the server runs on `0.0.0.0:8080`.

3.  **Test the endpoints:**
    *   **Ping server:** Open your browser or use a tool like `curl` to access:
        `http://localhost:8080/ping`
        You should receive:
        ```json
        {
            "message": "pong"
        }
        ```
    *   **Ping Database:** Open your browser or use a tool like `curl` to access:
        `http://localhost:8080/db-ping`
        If the database connection is successful, you should receive:
        ```json
        {
            "message": "Successfully pinged MongoDB!"
        }
        ```
        If there's an issue, you'll get an error message.

## Project Structure

-   `main.go`: The main application file, sets up Gin routes and initializes the database connection.
-   `db/db.go`: Contains the logic for connecting to MongoDB.
-   `.env`: Stores environment variables, including the MongoDB URI (not committed to version control).
-   `go.mod`, `go.sum`: Go module files managing project dependencies.
-   `README.md`: This file.

## How it Works

1.  When the application starts (`main.go`), it calls `db.ConnectDB()`.
2.  `db.ConnectDB()` in `db/db.go` loads environment variables from the `.env` file using `godotenv`.
3.  It retrieves the `MONGODB_URI`.
4.  It uses the MongoDB Go driver to establish a connection to the specified MongoDB instance.
5.  A ping to the primary node of the MongoDB cluster is performed to verify the connection.
6.  The `main.go` file sets up two routes:
    *   `/ping`: A simple health check for the Gin server.
    *   `/db-ping`: Pings the connected MongoDB database to check its status.

## Important Notes

*   Ensure your MongoDB server is running and accessible from where you're running the application.
*   If using a firewall, make sure the necessary ports (default 27017 for MongoDB) are open.
*   For MongoDB Atlas, ensure your IP address is whitelisted in the network access settings.
*   The `.env` file should **not** be committed to your version control system (e.g., Git). Add `.env` to your `.gitignore` file.
