package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db1 *sql.DB
var tmpl1 *template.Template

func main1() {
	var err error

	// MySQL connection
	db, err = sql.Open("mysql", "root:Zero2@tcp(localhost:3308)/gofrom")
	if err != nil {
		log.Fatal(err)
	}

	// Always test DB connection
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	tmpl = template.Must(template.ParseFiles("templates/form.html"))

	http.HandleFunc("/", formHandler)
	http.HandleFunc("/submit", submitHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	_, err := db.Exec(query, name, email)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, "Data saved successfully")
}
