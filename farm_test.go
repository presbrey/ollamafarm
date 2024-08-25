package ollamafarm_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/presbrey/ollamafarm"
)

func mustReadFile(filename string) string {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func setupTestServers() []*httptest.Server {
	servers := make([]*httptest.Server, 3)
	jsonFiles := []string{
		mustReadFile("tests/tags1.json"),
		mustReadFile("tests/tags2.json"),
		mustReadFile("tests/tags3.json"),
	}

	for i, jsonContent := range jsonFiles {
		content := jsonContent // Capture the content in a local variable
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/tags" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(content))
			} else if r.URL.Path == "/api/version" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"version":"0.3.6"}`))
			} else if r.URL.Path == "/" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
	}

	return servers
}

func TestFarmMethods(t *testing.T) {
	servers := setupTestServers()
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	farm := ollamafarm.New()
	farm2 := ollamafarm.NewWithOptions(&ollamafarm.Options{Heartbeat: time.Second, ModelsTTL: time.Second})

	// Register clients
	for i, server := range servers {
		err := farm.RegisterURL(server.URL, &ollamafarm.Properties{
			Group:    fmt.Sprintf("group%d", i+1),
			Priority: i + 1,
		})
		if err != nil {
			t.Fatalf("Failed to register client %d: %v", i+1, err)
		}
		farm2.RegisterURL(server.URL, nil)
		farm2.RegisterURL(server.URL, nil)
	}
	if farm2.RegisterURL("\000", nil) == nil {
		t.Error("Expected error")
	}

	// Wait for initial updates
	time.Sleep(time.Second)

	t.Run("TestSelect", func(t *testing.T) {
		llamas := farm.Select(&ollamafarm.Where{Group: "group0"})
		if len(llamas) != 0 {
			t.Errorf("Expected 0 client, got %d", len(llamas))
		}
		llamas = farm.Select(&ollamafarm.Where{Group: "group1"})
		if len(llamas) != 1 {
			t.Errorf("Expected 1 client, got %d", len(llamas))
		}
		llamas = farm.Select(&ollamafarm.Where{Model: "gemma2:27b-text-q6_K"})
		if len(llamas) != 2 {
			t.Errorf("Expected 2 llamas, got %d", len(llamas))
		}
		if llamas[0].Priority() != 1 {
			t.Errorf("Expected priority 1, got %d", llamas[0].Priority())
		}
		if llamas[1].Priority() != 2 {
			t.Errorf("Expected priority 2, got %d", llamas[1].Priority())
		}
		llamas = farm2.Select(nil)
		if len(llamas) != 3 {
			t.Errorf("Expected 3 llamas, got %d", len(llamas))
		}
		llamas = farm2.Select(&ollamafarm.Where{Model: "gemma2:27b-text-q6_K"})
		if len(llamas) != 2 {
			t.Errorf("Expected 2 llamas, got %d", len(llamas))
		}
	})

	t.Run("TestFirstByGroup", func(t *testing.T) {
		llama := farm.First(&ollamafarm.Where{Group: "group2"})
		if llama == nil {
			t.Error("Expected a llama, got nil")
		}
		if llama.Priority() != 2 {
			t.Errorf("Expected priority 2, got %d", llama.Priority())
		}
		if !llama.Online() {
			t.Error("Expected online, got offline")
		}
		llama = farm2.First(&ollamafarm.Where{Group: ""})
		if llama == nil {
			t.Error("Expected a llama, got nil")
		}
	})

	t.Run("TestFirstByModel", func(t *testing.T) {
		nilollama := farm.First(&ollamafarm.Where{Model: "do-not-find-nemo"})
		if nilollama != nil {
			t.Error("Expected a nil, got notnil")
		}
		nilollama2 := farm2.First(&ollamafarm.Where{Model: "do-not-find-nemo"})
		if nilollama2 != nil {
			t.Error("Expected a nil, got notnil")
		}
		ollama := farm.First(&ollamafarm.Where{Model: "mistral-nemo:12b-instruct-2407-fp16"})
		if ollama == nil {
			t.Error("Expected a llama, got nil")
		}
		ollama2 := farm2.First(&ollamafarm.Where{Model: "mistral-nemo:12b-instruct-2407-fp16"})
		if ollama2 == nil {
			t.Error("Expected a llama, got nil")
		}
		group := ollama.Group()
		if group != "group2" {
			t.Errorf("Expected group2, got %s", group)
		}
		version, _ := ollama.Client().Version(context.Background())
		if version != "0.3.6" {
			t.Errorf("Expected version 0.3.6, got %s", version)
		}
	})

	t.Run("TestModelCounts", func(t *testing.T) {
		models := farm.ModelCounts(nil)
		expectedModels := []string{
			"llama3.1:8b-instruct-q4_0",
			"llama3.1:8b-instruct-q8_0",
			"nomic-embed-text:latest",
			"phi3.5:3.8b-mini-instruct-q4_0",
			"phi3.5:3.8b-mini-instruct-q8_0",
			"starcoder2:3b",
		}
		if len(models) != 16 {
			t.Errorf("Expected %d models, got %d", 16, len(models))
		}
		for _, model := range expectedModels {
			found := false
			for m := range models {
				if m == model {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected model %s not found", model)
			}
		}
	})

	t.Run("TestOfflineServer", func(t *testing.T) {
		// Close the second server to simulate it going offline
		servers[1].Close()

		// Wait for the client to be marked as offline
		time.Sleep(6 * time.Second)

		llamas := farm.Select(&ollamafarm.Where{Group: "group2"})
		if len(llamas) != 0 {
			t.Errorf("Expected 0 llamas, got %d", len(llamas))
		}
		llamas = farm2.Select(nil)
		if len(llamas) != 2 {
			t.Errorf("Expected 2 llamas, got %d", len(llamas))
		}
	})

	t.Run("TestOllamaFarm", func(t *testing.T) {
		llama := farm.First(&ollamafarm.Where{Group: "group1"})
		if llama.Farm() != farm {
			t.Error("Expected farm, got nil")
		}
	})
}
