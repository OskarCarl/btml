package main

import (
	"flag"
	"log"
	"os"

	"github.com/vs-ude/btfl/internal/logging"
	"github.com/vs-ude/btfl/internal/model"
)

func main() {
	flag.Parse()

	logging.Logger.SetPrefix("[TEST-MODEL] ")
	logging.Logger.Use()

	// Create a new model instance
	m, err := model.NewSimpleModel(
		"model/",
		"python3",
		"data/prepared/fMNIST_train_split_42.pt",
		"data/prepared/fMNIST_test_split_42.pt",
		"logs/model.log",
	)
	if err != nil {
		log.Printf("Failed to create model: %v", err)
		os.Exit(1)
	}
	// Ensure cleanup on exit
	defer m.Close()

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
