package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv" // Import strconv for converting string to int

	"github.com/go-pg/pg/v10"
)

var db *pg.DB

// User struct to map to PostgreSQL users table
type User struct {
	ID    int
	Name  string
	Email string
}

// Initialize PostgreSQL database connection
func initDB() {
	db = pg.Connect(&pg.Options{
		User:     "postgres",       // Your PostgreSQL username
		Password: "N02070164",      // Your PostgreSQL password
		Database: "api",            // The database name you created in pgAdmin
		Addr:     "localhost:5432", // Default PostgreSQL port
	})

	// Test connection
	_, err := db.Exec("SELECT 1")
	if err != nil {
		log.Fatalf("Error connecting to database: %s\n", err)
	}
}

// Serve the index.html file
func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Get all users from the database and render them
func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	err := db.Model(&users).Select()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("users").Parse(`
		{{range .}}
			<li>{{.Name}} - {{.Email}} <a href="/delete/{{.ID}}">Delete</a></li>
		{{end}}
	`)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, users)
}

// Add a new user to the database
func addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")

		// Insert the new user into the database
		_, err := db.Model(&User{Name: name, Email: email}).Insert()
		if err != nil {
			http.Error(w, "Error adding user", http.StatusInternalServerError)
			return
		}

		// After adding a user, return the updated user list
		getUsers(w, r)
	}
}

// Delete a user from the database
func deleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path
	idStr := r.URL.Path[len("/delete/"):]

	// Convert the string ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Delete the user from the database
	_, err = db.Model(&User{ID: id}).Delete()
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	// After deletion, return the updated user list
	getUsers(w, r)
}

func main() {
	initDB()

	// Serve the index page
	http.HandleFunc("/", serveIndex)

	// API Endpoints for adding users and deleting users
	http.HandleFunc("/add", addUser)
	http.HandleFunc("/delete/", deleteUser)

	// Endpoint for fetching users
	http.HandleFunc("/users", getUsers)

	// Add styles.css file
	http.Handle("/style.css", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
