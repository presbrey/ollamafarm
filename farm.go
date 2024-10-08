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
func (f *Farm) RegisterClient(name string, client *api.Client, properties *Properties) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.ollamas[name]; exists {
		return
	}

	p := Properties{}
	if properties != nil {
		p.Group = properties.Group
		p.Offline = properties.Offline
		p.Priority = properties.Priority
	}

	ollama := &Ollama{
		name: name,

		client:     client,
		farm:       f,
		models:     make(map[string]*api.ListModelResponse),
		properties: p,
	}
	f.ollamas[name] = ollama

	go ollama.updateTickers()
}

// RegisterClient adds a new Ollama to the Farm if it doesn't already exist.
func (f *Farm) RegisterClientURL(name string, client *api.Client, properties *Properties, url *url.URL) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.ollamas[name]; exists {
		return
	}

	p := Properties{}
	if properties != nil {
		p.Group = properties.Group
		p.Offline = properties.Offline
		p.Priority = properties.Priority
	}

	ollama := &Ollama{
		name: name,
		url:  url,

		client:     client,
		farm:       f,
		models:     make(map[string]*api.ListModelResponse),
		properties: p,
	}
	f.ollamas[name] = ollama

	go ollama.updateTickers()
}

// RegisterNamedURL adds a new Ollama to the Farm using the given name as the ID.
func (f *Farm) RegisterNamedURL(name, baseURL string, properties *Properties) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	client := api.NewClient(parsedURL, http.DefaultClient)

	f.RegisterClientURL(name, client, properties, parsedURL)
	return nil
}

// RegisterURL adds a new Ollama to the Farm using the baseURL as the ID.
func (f *Farm) RegisterURL(baseURL string, properties *Properties) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	return f.RegisterNamedURL(parsedURL.String(), baseURL, properties)
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
	if where.Model != "" && ollama.models[where.Model] == nil {
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

// AllModels returns a list of all unique models available across all registered Ollamas.
func (f *Farm) AllModels() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	modelSet := make(map[string]struct{})
	for _, ollama := range f.ollamas {
		if !ollama.properties.Offline {
			for model := range ollama.models {
				modelSet[model] = struct{}{}
			}
		}
	}

	models := make([]string, 0, len(modelSet))
	for model := range modelSet {
		models = append(models, model)
	}

	return models
}
