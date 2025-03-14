package main

import (
	"log/slog"
	"os"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/model"
)

func main() {
	logging.FromEnv()

	c := &model.Config{
		Name:          "42",
		PythonRuntime: "python3",
		ModelPath:     "model/",
		Dataset:       "fMNIST",
		DataPath:      "data/prepared/",
		LogPath:       "logs/model.log",
	}
	// Create a new model instance
	m, err := model.NewModel(c, nil)
	if err != nil {
		slog.Error("Failed to create model", "error", err)
		os.Exit(1)
	}
	// Ensure cleanup on exit
	defer m.Shutdown()

	// Train the model
	if err := m.Train(); err != nil {
		slog.Error("Failed to train model", "error", err)
		os.Exit(1)
	}

	// Get initial weights
	weights, err := m.GetWeights()
	if err != nil {
		slog.Error("Failed to get weights", "error", err)
		os.Exit(1)
	}

	// Evaluate the model
	if err := m.Eval(); err != nil {
		slog.Error("Failed to evaluate model", "error", err)
		os.Exit(1)
	}

	// Apply weights back
	if err := m.Apply(weights); err != nil {
		slog.Error("Failed to apply weights", "error", err)
		os.Exit(1)
	}

	// Evaluate the model
	if err := m.Eval(); err != nil {
		slog.Error("Failed to evaluate model", "error", err)
		os.Exit(1)
	}
}
