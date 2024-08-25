package ollamafarm

import (
	"context"
	"time"

	"github.com/ollama/ollama/api"
)

// Client returns the Ollama client.
func (ollama *Ollama) Client() *api.Client {
	ollama.farm.mu.RLock()
	defer ollama.farm.mu.RUnlock()
	return ollama.client
}

// Group returns the Ollama's group.
func (ollama *Ollama) Group() string {
	ollama.farm.mu.RLock()
	defer ollama.farm.mu.RUnlock()
	return ollama.properties.Group
}

// Offline returns whether the Ollama is online.
func (ollama *Ollama) Online() bool {
	ollama.farm.mu.RLock()
	defer ollama.farm.mu.RUnlock()
	return !ollama.properties.Offline
}

// Priority returns the Ollama's priority.
func (ollama *Ollama) Priority() int {
	ollama.farm.mu.RLock()
	defer ollama.farm.mu.RUnlock()
	return ollama.properties.Priority
}

func (ollama *Ollama) updateModels() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	listResponse, err := ollama.client.List(ctx)
	cancel()

	ollama.farm.mu.Lock()
	if err != nil {
		ollama.properties.Offline = true
		ollama.models = make(map[string]bool)
	} else {
		ollama.properties.Offline = false
		for _, model := range listResponse.Models {
			ollama.models[model.Name] = true
		}
	}
	ollama.farm.mu.Unlock()
}

func (ollama *Ollama) updateOnline() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := ollama.client.Heartbeat(ctx)
	cancel()

	ollama.farm.mu.Lock()
	ollama.properties.Offline = err != nil
	ollama.farm.mu.Unlock()
}

// updateOllama fetches and updates the model list and checks the client's status.
func (ollama *Ollama) updateTickers() {
	ollama.farm.mu.Lock()
	heartbeatTicker := time.NewTicker(ollama.farm.options.Heartbeat)
	modelTicker := time.NewTicker(ollama.farm.options.ModelsTTL)
	ollama.farm.mu.Unlock()

	ollama.updateModels()
	ollama.updateOnline()

	for {
		select {
		case <-heartbeatTicker.C:
			ollama.updateOnline()
		case <-modelTicker.C:
			ollama.updateModels()
		}
	}
}
