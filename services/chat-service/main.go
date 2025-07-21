package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Message struct {
	ID           string    `json:"id"`
	Role         string    `json:"role"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}

type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "chat-service",
		})
	})

	http.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		conversations := []Conversation{
			{
				ID:        "1",
				Title:     "Sample Conversation",
				CreatedAt: time.Now(),
			},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"conversations": conversations,
		})
	})

	http.HandleFunc("/conversations/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path[len(r.URL.Path)-9:] == "/messages" {
			messages := []Message{
				{
					ID:        "1",
					Role:      "user",
					Content:   "Hello, how are you?",
					CreatedAt: time.Now(),
				},
				{
					ID:        "2",
					Role:      "assistant",
					Content:   "Hello! I am doing well, thank you for asking. How can I help you today?",
					CreatedAt: time.Now(),
				},
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"messages": messages,
			})
		}
	})

	log.Println("Chat service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
} 