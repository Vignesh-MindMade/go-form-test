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
		dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbSocket := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
	
	if dbUser == "" || dbName == "" || dbSocket == "" {
		log.Println("DB config missing — running WITHOUT database")
		return
	}

		dsn := fmt.Sprintf(
		"%s:%s@unix(%s)/%s?parseTime=true",
		dbUser, dbPass, dbSocket, dbName,
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Println("DB open failed:", err)
		db = nil
		return
	}

	if err := db.Ping(); err != nil {
		log.Println("DB ping failed:", err)
		db = nil
		return
	}

	log.Println("Database connected")
}
func main() {

	// Load .env file
	_ = godotenv.Load() // ignore error in Cloud Run
	os.MkdirAll("uploads", 0755)

	initDB()

	// Parse templates
	tmpl = template.Must(template.ParseFiles("templates/form.html"))

	http.HandleFunc("/", showForm)
	http.HandleFunc("/submit", submitForm)
	http.HandleFunc("/api/users", createUserAPI)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}

func showForm(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func submitForm(w http.ResponseWriter, r *http.Request) {

	// Max upload size = 10 MB
	r.ParseMultipartForm(50 << 20)

	// Text fields
	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	city := r.FormValue("city")

	// File: Image
	imageFile, imageHeader, _ := r.FormFile("image")
	defer imageFile.Close()

	imagePath := filepath.Join("uploads", imageHeader.Filename)
	saveFile(imageFile, imagePath)

	// File: PDF
	pdfFile, pdfHeader, _ := r.FormFile("pdf")
	defer pdfFile.Close()

	pdfPath := filepath.Join("uploads", pdfHeader.Filename)
	saveFile(pdfFile, pdfPath)

	// Insert into DB
	query := `
		INSERT INTO users
		(name, email, phone, city, image_path, pdf_path)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(
		query,
		name, email, phone, city,
		imagePath, pdfPath,
	)

	if err != nil {
		log.Println(err)
http.Error(w, "Internal error", 500)
return
	}

	fmt.Fprintln(w, "Data & files saved successfully")
}

// Helper function to save files
func saveFile(src io.Reader, path string) {
	dst, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	io.Copy(dst, src)
}

func createUserAPI(w http.ResponseWriter, r *http.Request) {

	// Allow only POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Content-Length:", r.ContentLength)
	// 🔒 HARD request limit (this is the real gatekeeper)
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form (memory buffer only)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Println("Multipart error:", err)
		http.Error(w, "File too large or invalid multipart data", http.StatusBadRequest)
		return
	}
	// Text fields
	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	city := r.FormValue("city")

	// Image
	imageFile, imageHeader, err := r.FormFile("image")
if err != nil {
	http.Error(w, "Image required", http.StatusBadRequest)
	return
}
defer imageFile.Close()

	imagePath := filepath.Join("uploads", imageHeader.Filename)
	saveFile(imageFile, imagePath)

	// PDF
	pdfFile, pdfHeader, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "PDF required", http.StatusBadRequest)
		return
	}
	defer pdfFile.Close()

	pdfPath := filepath.Join("uploads", pdfHeader.Filename)
	saveFile(pdfFile, pdfPath)

	// Insert DB
	query := `
		INSERT INTO users
		(name, email, phone, city, image_path, pdf_path)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query, name, email, phone, city, imagePath, pdfPath)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// API response (JSON)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"status":"success","message":"User created"}`)

}
