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

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var tmpl *template.Template

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read DB values from environment
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	tmpl = template.Must(template.ParseFiles("templates/form.html"))

	http.HandleFunc("/", showForm)
	http.HandleFunc("/submit", submitForm)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func showForm(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func submitForm(w http.ResponseWriter, r *http.Request) {

	// Max upload size = 10 MB
	r.ParseMultipartForm(10 << 20)

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
		log.Fatal(err)
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