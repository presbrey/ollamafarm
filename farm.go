package ollamafarm

import (
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/ollama/ollama/api"
)

// New creates a new Farm instance for managing for multiple Ollamas.
func New() *Farm {
	options := &Options{}
	return NewWithOptions(options)
}

// NewWithOptions creates a new Farm instance with the given options.
func NewWithOptions(options *Options) *Farm {
	if options.Client == nil {
		options.Client = http.DefaultClient
	}
	if options.Heartbeat == 0 {
		options.Heartbeat = 5 * time.Second
	}
	if options.ModelsTTL == 0 {
		options.ModelsTTL = 30 * time.Second
	}
	return &Farm{
		ollamas: make(map[string]*Ollama),
		options: options,
	}
}

// RegisterClient adds a new Ollama to the Farm if it doesn't already exist.
func (f *Farm) RegisterClient(id string, client *api.Client, properties *Properties) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.ollamas[id]; exists {
		return
	}

	p := Properties{}
	if properties != nil {
		p.Group = properties.Group
		p.Offline = properties.Offline
		p.Priority = properties.Priority
	}

	ollama := &Ollama{
		client:     client,
		farm:       f,
		models:     make(map[string]bool),
		properties: p,
	}
	f.ollamas[id] = ollama

	go ollama.updateTickers()
}

// RegisterURL adds a new Ollama to the Farm using the baseURL as the ID.
func (f *Farm) RegisterURL(baseURL string, properties *Properties) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	client := api.NewClient(parsedURL, http.DefaultClient)

	f.RegisterClient(parsedURL.String(), client, properties)
	return nil
}

// First returns the first Ollama that matches the given where.
func (f *Farm) First(where *Where) *Ollama {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var bestMatch *Ollama
	for _, ollama := range f.ollamas {
		if f.matchesWhere(ollama, where) {
			if bestMatch == nil || ollama.properties.Priority < bestMatch.properties.Priority {
				bestMatch = ollama
			}
		}
	}

	if bestMatch != nil {
		return bestMatch
	}
	return nil
}

// Select returns a list of Ollamas that match the given where, sorted by ascending Priority.
func (f *Farm) Select(where *Where) []*Ollama {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var matches []*Ollama
	for _, ollama := range f.ollamas {
		if f.matchesWhere(ollama, where) {
			matches = append(matches, ollama)
		}
	}

	// Sort matches by ascending Priority
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].properties.Priority < matches[j].properties.Priority
	})

	return matches
}

// matchesWhere checks if an Ollama matches the given where.
func (f *Farm) matchesWhere(ollama *Ollama, where *Where) bool {
	if where == nil {
		return !ollama.properties.Offline
	}
	if where.Group != "" && ollama.properties.Group != where.Group {
		return false
	}
	if where.Model != "" && !ollama.models[where.Model] {
		return false
	}
	if where.Offline != ollama.properties.Offline {
		return false
	}
	return true
}

// ModelCounts returns a count of all models available across all registered Ollamas.
func (f *Farm) ModelCounts(where *Where) map[string]uint {
	f.mu.RLock()
	defer f.mu.RUnlock()

	modelCounts := make(map[string]uint)
	for _, ollama := range f.ollamas {
		if where == nil || f.matchesWhere(ollama, where) {
			if !ollama.properties.Offline {
				for model := range ollama.models {
					modelCounts[model] += 1
				}
			}
		}
	}

	return modelCounts
}
