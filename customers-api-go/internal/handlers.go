package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func HandleCreateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Customer
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		res, err := db.Exec("INSERT INTO customers (name, email, phone) VALUES (?, ?, ?)", c.Name, c.Email, c.Phone)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		c.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func HandleListCustomers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email, phone FROM customers")
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var customers []Customer
		for rows.Next() {
			var c Customer
			if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone); err != nil {
				http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			customers = append(customers, c)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customers)
	}
}

func HandleGetCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.Atoi(idStr)

		var c Customer
		err := db.QueryRow("SELECT id, name, email, phone FROM customers WHERE id = ?", id).
			Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func HandleUpdateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.Atoi(idStr)

		var c Customer
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE customers SET name=?, email=?, phone=? WHERE id=?", c.Name, c.Email, c.Phone, id)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		c.ID = id

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func HandleDeleteCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.Atoi(idStr)

		_, err := db.Exec("DELETE FROM customers WHERE id=?", id)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"deleted"}`))
	}
}
