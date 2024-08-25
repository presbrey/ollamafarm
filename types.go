package ollamafarm

import (
	"net/http"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
)

// Farm manages multiple Ollamas.
type Farm struct {
	ollamas map[string]*Ollama
	options *Options

	mu sync.RWMutex
}

// Ollama stores information about an Ollama server.
type Ollama struct {
	client     *api.Client
	farm       *Farm
	models     map[string]bool
	properties Properties
}

// Options defines the options for an Farm.
type Options struct {
	// Client is the HTTP client used to make requests.
	Client *http.Client
	// ModelsTTL is the time-to-live for the models cache. Default: 30 seconds
	ModelsTTL time.Duration
	// PingTTL is the time-to-live for the online/offline ping. Default: 5 seconds
	PingTTL time.Duration
}

// Properties defines the properties of an Ollama client.
type Properties struct {
	Group    string
	Offline  bool
	Priority int
}

// Where defines the selection criteria for Ollama clients.
type Where struct {
	Group   string
	Model   string
	Offline bool
}
