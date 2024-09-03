package main

import (
	"log"
	"net/http"

	"github.com/presbrey/ollamafarm"
	"github.com/presbrey/ollamafarm/server"
)

func main() {
	farm := ollamafarm.New()

	// Register your Ollama clients here
	// For example:
	// farm.RegisterURL("http://localhost:11434", nil)

	s := server.NewServer(farm)

	http.HandleFunc("/version", s.VersionHandler)
	http.HandleFunc("/models", s.ModelsHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
