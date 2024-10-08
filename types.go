package ollamafarm

import (
	"net/http"
	"net/url"
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
	name string
	url  *url.URL

	client     *api.Client
	farm       *Farm
	models     map[string]*api.ListModelResponse
	properties Properties
}

// Options defines the options for an Farm.
type Options struct {
	// Client is the HTTP client used to make requests.
	Client *http.Client
	// Heartbeat is the time-to-live for online/offline detection. Default: 5 seconds
	Heartbeat time.Duration
	// ModelsTTL is the time-to-live for the models cache. Default: 30 seconds
	ModelsTTL time.Duration
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
