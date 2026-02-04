package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB
var tmpl *template.Template

const maxUploadSize = 200 << 20 // 200 MB

func initDB() {
	dbUser := os.Getenv("_DB_USER")
	dbPass := os.Getenv("_DB_PASS")
	dbHost := os.Getenv("_DB_HOST")
	dbPort := os.Getenv("_DB_PORT")
	dbName := os.Getenv("_DB_NAME")

	if dbUser == "" || dbHost == "" || dbName == "" {
		log.Println("WARNING: DB environment variables missing → running WITHOUT database")
		return
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("ERROR: cannot open database connection: %v", err)
		db = nil
		return
	}

	if err = db.Ping(); err != nil {
		log.Printf("ERROR: database ping failed: %v → disabling DB", err)
		db = nil
		return
	}

	log.Println("INFO: Database connected successfully")
}

func main() {
	// Load .env file (ignore error in Cloud Run / container environments)
	_ = godotenv.Load()

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Printf("WARNING: cannot create uploads directory: %v", err)
		// continue anyway — uploads will fail later if needed
	}

	initDB()

	var err error
	tmpl, err = template.ParseFiles("templates/form.html")
	if err != nil {
		log.Fatalf("CRITICAL: cannot parse template: %v", err) // this one is usually fatal
		// If you want even this non-fatal → comment out log.Fatal and return / os.Exit(1)
	}

	// Register handlers
	http.HandleFunc("/", showForm)
	http.HandleFunc("/submit", submitForm)
	http.HandleFunc("/api/users", createUserAPI)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on :%s ...", port)

	// Most important change: do NOT use log.Fatal here
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("ERROR: HTTP server failed: %v", err)
		// You can decide what to do:
		// Option A: just log and let program continue (if you have other goroutines)
		// Option B: graceful shutdown or os.Exit(1) — but only if this is the main purpose
		//
		// Recommended for most cloud/container cases:
		log.Println("Server stopped. Exiting.")
		os.Exit(1) // ← explicit exit is clearer than log.Fatal
	}
}

func showForm(w http.ResponseWriter, r *http.Request) {
	if tmpl == nil {
		http.Error(w, "Template not loaded", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func submitForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		log.Printf("ParseMultipartForm failed: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	city := r.FormValue("city")

	// Image (optional in this handler?)
	imageFile, imageHeader, _ := r.FormFile("image")
	var imagePath string
	if imageFile != nil {
		defer imageFile.Close()
		imagePath = filepath.Join("uploads", imageHeader.Filename)
		if err := saveFile(imageFile, imagePath); err != nil {
			log.Printf("Failed to save image: %v", err)
		}
	}

	// PDF (optional?)
	pdfFile, pdfHeader, _ := r.FormFile("pdf")
	var pdfPath string
	if pdfFile != nil {
		defer pdfFile.Close()
		pdfPath = filepath.Join("uploads", pdfHeader.Filename)
		if err := saveFile(pdfFile, pdfPath); err != nil {
			log.Printf("Failed to save pdf: %v", err)
		}
	}

	if db == nil {
		fmt.Fprintln(w, "Data & files saved (but DB is not connected)")
		return
	}

	query := `INSERT INTO users (name, email, phone, city, image_path, pdf_path) VALUES (?,?,?,?,?,?)`
	_, err := db.Exec(query, name, email, phone, city, imagePath, pdfPath)
	if err != nil {
		log.Printf("DB insert failed: %v", err)
		http.Error(w, "Internal server error (database)", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Data & files saved successfully")
}

func saveFile(src io.Reader, path string) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func createUserAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Parse multipart error: %v", err)
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	city := r.FormValue("city")

	imageFile, imageHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer imageFile.Close()

	imagePath := filepath.Join("uploads", imageHeader.Filename)
	if err := saveFile(imageFile, imagePath); err != nil {
		log.Printf("Failed to save image: %v", err)
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	pdfFile, pdfHeader, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "PDF file is required", http.StatusBadRequest)
		return
	}
	defer pdfFile.Close()

	pdfPath := filepath.Join("uploads", pdfHeader.Filename)
	if err := saveFile(pdfFile, pdfPath); err != nil {
		log.Printf("Failed to save pdf: %v", err)
		http.Error(w, "Failed to save PDF", http.StatusInternalServerError)
		return
	}

	if db == nil {
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	query := `INSERT INTO users (name, email, phone, city, image_path, pdf_path) VALUES (?,?,?,?,?,?)`
	_, err = db.Exec(query, name, email, phone, city, imagePath, pdfPath)
	if err != nil {
		log.Printf("Database insert failed: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"status":"success","message":"User created"}`)
}