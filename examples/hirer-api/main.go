// Hirer API - Example Service
// This is a simple Go HTTP service demonstrating cross-domain integration

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Job represents a job posting
type Job struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Description string    `json:"description"`
	Skills      []string  `json:"skills"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Response is a generic API response
type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var jobs = []Job{
	{ID: "1", Title: "Senior Backend Engineer", Company: "TechCorp", Description: "Building scalable systems", Skills: []string{"Go", "Kubernetes"}, CreatedAt: time.Now()},
	{ID: "2", Title: "ML Engineer", Company: "AIStartup", Description: "Developing ML models", Skills: []string{"Python", "TensorFlow"}, CreatedAt: time.Now()},
	{ID: "3", Title: "Platform Engineer", Company: "CloudInc", Description: "Building internal platform", Skills: []string{"Kubernetes", "Terraform"}, CreatedAt: time.Now()},
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
	http.HandleFunc("/api/v1/jobs", jobsHandler)
	http.HandleFunc("/api/v1/jobs/", jobByIDHandler)
	http.HandleFunc("/api/v1/match", matchCandidatesHandler)

	log.Printf("Starting Hirer API on port %s", port)
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
		Message: "Hirer API v1.0.0 - XYZ Platform",
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

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(Response{Status: "ok", Data: jobs})
	case http.MethodPost:
		var newJob Job
		if err := json.NewDecoder(r.Body).Decode(&newJob); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newJob.ID = fmt.Sprintf("%d", len(jobs)+1)
		newJob.CreatedAt = time.Now()
		jobs = append(jobs, newJob)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response{Status: "created", Data: newJob})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func jobByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Path[len("/api/v1/jobs/"):]

	for _, j := range jobs {
		if j.ID == id {
			json.NewEncoder(w).Encode(Response{Status: "ok", Data: j})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Status: "error", Message: "Job not found"})
}

// matchCandidatesHandler demonstrates cross-domain integration
// It calls the Candidate API to find matching candidates for a job
func matchCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the Candidate API URL from environment or use default
	candidateAPIURL := os.Getenv("CANDIDATE_API_URL")
	if candidateAPIURL == "" {
		candidateAPIURL = "http://candidate-api.candidate.svc.cluster.local/api/v1/candidates"
	}

	// Call Candidate API (demonstrating cross-domain integration)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(candidateAPIURL)
	if err != nil {
		log.Printf("Error calling Candidate API: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Unable to reach Candidate API",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "error", Message: "Error reading response"})
		return
	}

	// Return the candidates data
	w.Write(body)
}

