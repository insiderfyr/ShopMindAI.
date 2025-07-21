package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "user-service",
		})
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		users := []User{
			{ID: "1", Username: "demo", Email: "demo@example.com"},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
		})
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := r.URL.Path[len("/users/"):]
		user := User{
			ID:       id,
			Username: "demo",
			Email:    "demo@example.com",
		}
		json.NewEncoder(w).Encode(user)
	})

	log.Println("User service starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}