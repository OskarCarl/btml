package main

import (
	"flag"
	"log"
	"os"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/model"
)

func main() {
	flag.Parse()

	logging.Logger.SetPrefix("[TEST-MODEL] ")
	logging.Logger.Use()

	c := &model.Config{
		Name:          "42",
		PythonRuntime: "python3",
		ModelPath:     "model/",
		Dataset:       "fMNIST",
		DataPath:      "data/prepared/",
		LogPath:       "logs/model.log",
	}
	// Create a new model instance
	m, err := model.NewSimpleModel(c)
	if err != nil {
		log.Printf("Failed to create model: %v", err)
		os.Exit(1)
	}
	// Ensure cleanup on exit
	defer m.Shutdown()

	// Train the model
	if err := m.Train(); err != nil {
		log.Printf("Failed to train model: %v", err)
		os.Exit(1)
	}

	// Get initial weights
	weights, err := m.GetWeights()
	if err != nil {
		log.Printf("Failed to get weights: %v", err)
		os.Exit(1)
	}

	// Evaluate the model
	if err := m.Eval(); err != nil {
		log.Printf("Failed to evaluate model: %v", err)
		os.Exit(1)
	}

	// Apply weights back
	if err := m.Apply(weights); err != nil {
		log.Printf("Failed to apply weights: %v", err)
		os.Exit(1)
	}

	// Evaluate the model
	if err := m.Eval(); err != nil {
		log.Printf("Failed to evaluate model: %v", err)
		os.Exit(1)
	}
}
