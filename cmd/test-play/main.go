package main

import (
	"fmt"
	"os"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
	"github.com/vs-ude/btml/internal/play"
)

func main() {
	mconf := &model.Config{
		Name:          "42",
		PythonRuntime: "venv/bin/python3",
		ModelPath:     "model",
		DataPath:      "data/prepared",
		LogPath:       "",
		Dataset:       "fMNIST",
	}
	mod, err := model.NewModel(mconf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	me := peer.NewMe(&peer.Config{})
	p := play.NewPlay(me, mod)
	p.AddStep(&play.Train{})
	p.AddStep(&play.Eval{})
	p.AddStep(&play.Wait{T: time.Second * 10})
	out, err := p.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println(string(out))

	err = mod.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	err = p.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
}
