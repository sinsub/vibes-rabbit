package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database connection
	initDB()

	// Initialize router
	http.HandleFunc("/api/v1/hello", helloHandler)
	http.HandleFunc("/api/v1/messages", messagesHandler)

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDB() {
	// Get database connection details from environment
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	log.Println("Database connected successfully")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"hello": "world",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getMessages(w, r)
	case "POST":
		createMessage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getMessages(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT id, message, created_at FROM messages ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id int
		var message string
		var createdAt time.Time

		err := rows.Scan(&id, &message, &createdAt)
		if err != nil {
			continue
		}

		messages = append(messages, map[string]interface{}{
			"id":         id,
			"message":    message,
			"created_at": createdAt,
		})
	}

	response := map[string]interface{}{
		"messages": messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	var id int
	err := db.QueryRow("INSERT INTO messages (message) VALUES ($1) RETURNING id", request.Message).Scan(&id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"id":      id,
		"message": request.Message,
		"status":  "created",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
