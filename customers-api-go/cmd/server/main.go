package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

type Customer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func main() {
	// Get env vars from Railway
	dbUser := os.Getenv("MYSQLUSER")
	dbPass := os.Getenv("MYSQLPASSWORD")
	dbHost := os.Getenv("MYSQLHOST")
	dbPort := os.Getenv("MYSQLPORT")
	dbName := os.Getenv("MYSQLDATABASE")

	// Fallback for local dev (XAMPP)
	if dbUser == "" {
		dbUser = "root"
	}
	if dbPass == "" {
		dbPass = "" // XAMPP default = empty
	}
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "testdb"
	}

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	// Ping test
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	log.Println("✅ Successfully connected to MySQL!")

	// ---- Router ----
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/customers", handleCreateCustomer(db))
	r.Get("/customers", handleListCustomers(db))
	r.Get("/customers/{id}", handleGetCustomer(db))
	r.Put("/customers/{id}", handleUpdateCustomer(db))
	r.Delete("/customers/{id}", handleDeleteCustomer(db))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// ---- Correct port ----
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func handleCreateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Customer
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		res, err := db.Exec("INSERT INTO customers (name,email,phone) VALUES (?,?,?)", c.Name, c.Email, c.Phone)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, _ := res.LastInsertId()
		c.ID = int(id)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func handleListCustomers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,name,email FROM customers")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var customers []Customer
		for rows.Next() {
			var c Customer
			rows.Scan(&c.ID, &c.Name, &c.Email)
			customers = append(customers, c)
		}
		json.NewEncoder(w).Encode(customers)
	}
}

func handleGetCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var c Customer
		err := db.QueryRow("SELECT id,name,email,phone,created_at,updated_at FROM customers WHERE id=?", id).
			Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.CreatedAt, &c.UpdatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(c)
	}
}

func handleUpdateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var c Customer
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE customers SET name=?,email=?,phone=? WHERE id=?", c.Name, c.Email, c.Phone, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c.ID, _ = strconv.Atoi(id)
		json.NewEncoder(w).Encode(c)
	}
}

func handleDeleteCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		_, err := db.Exec("DELETE FROM customers WHERE id=?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Customer deleted"})
	}
}
