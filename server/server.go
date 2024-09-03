package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/presbrey/ollamafarm"
)

// Server is an HTTP server that proxies requests to Ollamas on a Farm.
type Server struct {
	farm *ollamafarm.Farm
	mux  *http.ServeMux
}

// NewServer creates a new Server instance with the given Farm.
func NewServer(farm *ollamafarm.Farm) *Server {
	s := &Server{farm: farm}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", s.handleTags)
	mux.HandleFunc("/api/version", s.handleVersion)
	mux.HandleFunc("/", s.catchAllPost)
	s.mux = mux
	return s
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	ollama := s.farm.First(nil)
	if ollama == nil {
		http.Error(w, "No available Ollama instances", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	version, err := ollama.Client().Version(ctx)
	if err != nil {
		http.Error(w, "Failed to get version", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": version})
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	models := s.farm.AllModels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func (s *Server) catchAllPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	model, ok := body["model"].(string)
	if !ok {
		http.Error(w, "Missing or invalid 'model' field", http.StatusBadRequest)
		return
	}

	ollama := s.farm.First(&ollamafarm.Where{Model: model})
	if ollama == nil {
		http.Error(w, "No available Ollama instance for the specified model", http.StatusServiceUnavailable)
		return
	}

	// Create a new request to the selected Ollama
	proxyURL := ollama.BaseURL().ResolveReference(r.URL)
	proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send the request to the Ollama instance
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Error proxying request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		// We've already started writing the response, so we can't use http.Error here
		// Just log the error
		log.Printf("Error copying response body: %v", err)
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
