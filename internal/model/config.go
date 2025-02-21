package model

import (
	"fmt"
	"path"
)

type Config struct {
	Name          string
	PythonRuntime string
	ModelPath     string
	DataPath      string
	LogPath       string
	Dataset       string
}

func (c *Config) GetTrainDataPath() string {
	return path.Clean(fmt.Sprintf("%s/%s_train_split_%s.pt", c.DataPath, c.Dataset, c.Name))
}

func (c *Config) GetTestDataPath() string {
	return path.Clean(fmt.Sprintf("%s/%s_test_split_%s.pt", c.DataPath, c.Dataset, c.Name))
}
