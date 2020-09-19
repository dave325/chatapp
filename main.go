package main

import (
	"database/sql"
	database "database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type user struct {
	id        int
	username  string
	password  string
	createdAt time.Time
}

func createTable(db *database.DB) {
	query := `
    CREATE TABLE users (
        id INT AUTO_INCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        created_at DATETIME,
        PRIMARY KEY (id)
    );`

	// Executes the SQL query in our database. Check err to ensure there was no error.
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertUser(db *database.DB) {
	var (
		id        int
		username  string
		password  string
		createdAt time.Time
	)

	username = "johndoe"
	password = "secret"
	createdAt = time.Now()

	// Inserts our data into the users table and returns with the result and a possible error.
	// The result contains information about the last inserted id (which was auto-generated for us) and the count of rows this query affected.
	result, err := db.Exec(`INSERT INTO users (username, password, created_at) VALUES (?, ?, ?)`, username, password, createdAt)
	if err != nil {
		log.Fatal(err)
	}

	insertedID, error := result.LastInsertId()
	if error != nil {
		log.Fatal(error)
	}
	fmt.Println(insertedID)

	// Query the database and scan the values into out variables. Don't forget to check for errors.
	query := "SELECT id, username, password, created_at FROM users WHERE id = ?"
	if err := db.QueryRow(query, 1).Scan(&id, &username, &password, &createdAt); err != nil {
		log.Fatal(err)
	}

	fmt.Println(id, username, password, createdAt)
}

func main() {
	log.Print("Starting")
	// Configure the database connection (always check errors)
	db, err := sql.Open("mysql", "root:password@(127.0.0.1:3306)/test?parseTime=true")

	// Initialize the first connection to the database, to see if everything works correctly.
	// Make sure to check the error.
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	createTable(db)
	insertUser(db)

	r := mux.NewRouter()

	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		// get the book
		// navigate to the page
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]

		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})

	/*
		Path Prefixes and Subrouters - Useful when we want to create CRUD with a specific prefix
		bookrouter := r.PathPrefix("/books").Subrouter()
		bookrouter.HandleFunc("/", AllBooks)
		bookrouter.HandleFunc("/{title}", GetBook)


	*/
	http.ListenAndServe(":8001", r)

}
