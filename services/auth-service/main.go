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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "auth-service",
		})
	})

	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var login LoginRequest
		json.NewDecoder(r.Body).Decode(&login)

		if login.Username == "demo" && login.Password == "demo123" {
			response := LoginResponse{
				Token: "mock-jwt-token-12345",
				User: User{
					ID:       "1",
					Username: "demo",
					Email:    "demo@example.com",
				},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid credentials",
			})
		}
	})

	http.HandleFunc("/auth/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user := User{
			ID:       "1",
			Username: "demo",
			Email:    "demo@example.com",
		}
		json.NewEncoder(w).Encode(user)
	})

	log.Println("Auth service starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
} 