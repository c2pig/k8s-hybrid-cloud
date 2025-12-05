// Candidate API - Example Service
// This is a simple Go HTTP service demonstrating a typical tenant application

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Candidate represents a job candidate
type Candidate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Skills    []string  `json:"skills"`
	CreatedAt time.Time `json:"createdAt"`
}

// Response is a generic API response
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var candidates = []Candidate{
	{ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Skills: []string{"Go", "Kubernetes", "AWS"}, CreatedAt: time.Now()},
	{ID: "2", Name: "Bob Smith", Email: "bob@example.com", Skills: []string{"Python", "ML", "TensorFlow"}, CreatedAt: time.Now()},
	{ID: "3", Name: "Carol Williams", Email: "carol@example.com", Skills: []string{"Java", "Spring", "PostgreSQL"}, CreatedAt: time.Now()},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/api/v1/candidates", candidatesHandler)
	http.HandleFunc("/api/v1/candidates/", candidateByIDHandler)

	log.Printf("Starting Candidate API on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, logRequest(http.DefaultServeMux)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(Response{
		Status:  "ok",
		Message: "Candidate API v1.0.0 - XYZ Platform",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "healthy"})
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "ready"})
}

func candidatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(Response{Status: "ok", Data: candidates})
	case http.MethodPost:
		var newCandidate Candidate
		if err := json.NewDecoder(r.Body).Decode(&newCandidate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newCandidate.ID = fmt.Sprintf("%d", len(candidates)+1)
		newCandidate.CreatedAt = time.Now()
		candidates = append(candidates, newCandidate)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response{Status: "created", Data: newCandidate})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func candidateByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Path[len("/api/v1/candidates/"):]

	for _, c := range candidates {
		if c.ID == id {
			json.NewEncoder(w).Encode(Response{Status: "ok", Data: c})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Status: "error", Message: "Candidate not found"})
}

