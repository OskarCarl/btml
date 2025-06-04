package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/vs-ude/btml/internal/logging"
	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
	"github.com/vs-ude/btml/internal/play"
	"github.com/vs-ude/btml/internal/structs"
)

func main() {
	logging.FromEnv()

	mconf := &model.Config{
		Name:          "42",
		PythonRuntime: "venv/bin/python3",
		ModelPath:     "model",
		DataPath:      "data/prepared",
		LogPath:       "",
		Dataset:       "fMNIST",
	}
	mod, err := model.NewModel(mconf, nil)
	if err != nil {
		slog.Error("Failed to create model", "error", err)
		os.Exit(1)
	}
	me := peer.NewMe(&peer.Config{}, nil, &structs.Peer{})
	p := play.NewPlay(me, mod)
	p.AddStep(&play.Train{})
	p.AddStep(&play.Eval{})
	p.AddStep(&play.Wait{T: time.Second * 10})
	out, err := p.MarshalJSON()
	if err != nil {
		slog.Error("Failed to marshal play", "error", err)
		os.Exit(2)
	}
	fmt.Println(string(out))

	err = mod.Start()
	if err != nil {
		slog.Error("Failed to start model", "error", err)
		os.Exit(3)
	}

	err = p.Run()
	if err != nil {
		slog.Error("Failed to run play", "error", err)
		os.Exit(4)
	}
}
