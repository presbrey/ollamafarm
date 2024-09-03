package server

import (
	"encoding/json"
	"net/http"

	"github.com/presbrey/ollamafarm"
)

type Server struct {
	Farm *ollamafarm.Farm
}

func NewServer(farm *ollamafarm.Farm) *Server {
	return &Server{Farm: farm}
}

func (s *Server) VersionHandler(w http.ResponseWriter, r *http.Request) {
	ollama := s.Farm.First(nil)
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

func (s *Server) ModelsHandler(w http.ResponseWriter, r *http.Request) {
	models := s.Farm.AllModels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}
