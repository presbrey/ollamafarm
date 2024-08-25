# OllamaFarm

OllamaFarm is a Go package that manages multiple Ollama instances, providing a convenient way to interact with a farm of Ollama servers. It offers features like client registration, server selection based on criteria, and model availability tracking.

## Installation

To install OllamaFarm, use the following command:

```bash
go get github.com/presbrey/ollamafarm
```

## Usage

Here's a basic example of how to use OllamaFarm:

```go
package main

import (
    "fmt"
    "github.com/presbrey/ollamafarm"
)

func main() {
    farm := ollamafarm.New()

    // Register Ollama servers
    farm.RegisterURL("http://ollama1:11434", &ollamafarm.Properties{Group: "group1", Priority: 1})
    farm.RegisterURL("http://ollama2:11434", &ollamafarm.Properties{Group: "group2", Priority: 2})

    // Select an Ollama instance
    ollama := farm.First(&ollamafarm.Where{Group: "group1"})
    if ollama != nil {
        fmt.Printf("Selected Ollama from group: %s\n", ollama.Group())
    }

    // Get model counts
    modelCounts := farm.ModelCounts(nil)
    fmt.Printf("Available models: %v\n", modelCounts)
}
```

## API Reference

### Types

#### Farm

The main struct that manages multiple Ollama instances.

```go
type Farm struct {
    // unexported fields
}
```

#### Ollama

Represents an individual Ollama server.

```go
type Ollama struct {
    // unexported fields
}
```

#### Options

Defines the options for a Farm. All fields are optional.

```go
type Options struct {
    Client     *http.Client
    ModelsTTL  time.Duration
    PingTTL    time.Duration
}
```

#### Properties

Defines the properties of an Ollama client. All fields are optional.

```go
type Properties struct {
    Group    string
    Offline  bool
    Priority int
}
```

#### Where

Defines the selection criteria for Ollama clients.

```go
type Where struct {
    Group   string
    Model   string
    Offline bool
}
```

### Functions

#### New() *Farm

Creates a new Farm instance with default options.

#### NewWithOptions(options *Options) *Farm

Creates a new Farm instance with the given options.

### Farm Methods

#### RegisterClient(id string, client *api.Client, properties *Properties)

Adds a new Ollama to the Farm if it doesn't already exist.

#### RegisterURL(baseURL string, properties *Properties) error

Adds a new Ollama to the Farm using the baseURL as the ID.

#### First(where *Where) *Ollama

Returns the first Ollama that matches the given criteria.

#### Select(where *Where) []*Ollama

Returns a list of Ollamas that match the given criteria, sorted by ascending Priority.

#### ModelCounts(where *Where) map[string]uint

Returns a count of all models available across all registered Ollamas.

### Ollama Methods

#### Client() *api.Client

Returns the Ollama client.

#### Group() string

Returns the Ollama's group.

#### Online() bool

Returns whether the Ollama is online.

#### Priority() int

Returns the Ollama's priority.

## Contributing

Contributions to OllamaFarm are welcome! Please note the following guidelines:

1. All pull requests must maintain or improve the existing test coverage.
2. New features or changes must not break any existing APIs.
3. Write clear, concise commit messages.
4. Follow Go best practices and style guidelines.

## License

This project is licensed under the [LICENSE](LICENSE.txt) file in the root directory of this repository.
