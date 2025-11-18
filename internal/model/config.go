package model

import (
	"fmt"
	"os"
	"path"

	"github.com/google/shlex"
)

type Config struct {
	Name          string
	PythonRuntime string
	ModelArgs     []string
	DataPath      string
	LogPath       string
	Dataset       string
}

func (c *Config) GetTrainDataPath() string {
	return path.Clean(fmt.Sprintf("%s/%s/train_split_%s.pt", c.DataPath, c.Dataset, c.Name))
}

func (c *Config) GetTestDataPath() string {
	return path.Clean(fmt.Sprintf("%s/%s/test_split_%s.pt", c.DataPath, c.Dataset, c.Name))
}

func (c *Config) GetCheckpointPath() string {
	return path.Clean(fmt.Sprintf("%s/checkpoints/%s_%s", c.DataPath, c.Dataset, c.Name))
}

func FromEnv() *Config {
	line := os.Getenv("PYTHON_MODEL_LINE")
	c := &Config{
		Name:          "0",
		PythonRuntime: ".venv/bin/python3",
		ModelArgs:     []string{"model/main.py"},
		DataPath:      "model/data",
		LogPath:       "logs",
		Dataset:       "fMNIST",
	}
	if line != "" {
		f, err := shlex.Split(line)
		if err != nil {
			fmt.Printf("Error splitting PYTHON_MODEL_LINE: %v\n", err)
			return c
		}
		c.PythonRuntime = f[0]
		if len(f) > 1 {
			c.ModelArgs = f[1:]
		}
	}
	return c
}
