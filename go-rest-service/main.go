package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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
	http.HandleFunc("/api/v1/files", filesHandler)
	http.HandleFunc("/api/v1/files/", fileDownloadHandler)

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

// POST /api/v1/files: upload a .json file with a name
// GET  /api/v1/files: list all files (id, name, created_at)
func filesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		uploadFile(w, r)
	case "GET":
		listFiles(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	fileHeader, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer fileHeader.Close()

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Only allow .json files
	if !strings.HasSuffix(strings.ToLower(name), ".json") {
		http.Error(w, "File name must end with .json", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(fileHeader)
	if err != nil {
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	var id int
	err = db.QueryRow("INSERT INTO files (name, file) VALUES ($1, $2) RETURNING id", name, fileBytes).Scan(&id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"id":     id,
		"name":   name,
		"status": "uploaded",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func listFiles(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT id, name, created_at FROM files ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []map[string]any
	for rows.Next() {
		var id int
		var name string
		var createdAt time.Time
		err := rows.Scan(&id, &name, &createdAt)
		if err != nil {
			continue
		}
		files = append(files, map[string]any{
			"id":         id,
			"name":       name,
			"created_at": createdAt,
		})
	}
	response := map[string]any{"files": files}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /api/v1/files/{id}: download the file
func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract id from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid file id", http.StatusBadRequest)
		return
	}
	idStr := parts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid file id", http.StatusBadRequest)
		return
	}

	var name string
	var fileBytes []byte
	err = db.QueryRow("SELECT name, file FROM files WHERE id = $1", id).Scan(&name, &fileBytes)
	if err == sql.ErrNoRows {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(path.Base(name)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}
