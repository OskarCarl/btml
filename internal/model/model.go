package model

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Model interface {
	Eval() error
	Train() error
	Apply(Weights) error
	GetWeights() (Weights, error)
	Shutdown()
}

type SimpleModel struct {
	client  *ModelClient
	age     int
	command *exec.Cmd
}

func (m *SimpleModel) Shutdown() {
	m.client.Close()
	m.command.Process.Signal(syscall.SIGTERM)
	m.command.Wait()
	log.Default().Println("Model stopped")
}

func (m *SimpleModel) Eval() error {
	met, err := m.client.Eval()
	if err != nil {
		return fmt.Errorf("failed to evaluate model: %w", err)
	}
	log.Default().Printf("Evaluated model; acc: %f, loss: %f", met.acc, met.loss)
	return nil
}

func (m *SimpleModel) Train() error {
	met, err := m.client.Train()
	if err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}
	m.age++
	log.Default().Printf("Trained model to age %d; loss: %f", m.age, met.loss)
	return nil
}

func (m *SimpleModel) Apply(weights Weights) error {
	ratio := getRatio(m, weights)
	if err := m.client.Apply(weights, ratio); err != nil {
		return fmt.Errorf("failed to apply weights to model: %w", err)
	}
	log.Default().Print("Applied weights to model")
	updateAge(m, weights)
	return nil
}

func (m *SimpleModel) GetWeights() (Weights, error) {
	w, err := m.client.GetWeights()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weights from model: %w", err)
	}
	log.Default().Print("Got weights from model")
	w.setAge(m.age)
	return w, nil
}

// NewModel creates a new Model instance by starting the Python process
// and establishing a connection to it
func NewSimpleModel(c *Config) (Model, error) {
	// Create a random socket path in /tmp
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("btml-model-%d.sock", time.Now().Unix()))

	// Start the Python process
	args := []string{
		"main.py",
		"--train-data", c.GetTrainDataPath(),
		"--test-data", c.GetTestDataPath(),
		"--socket", socketPath,
	}
	if c.LogPath != "" {
		if p, err := resolveLogPath(c); err == nil {
			args = append(args, "--log-file", p)
		} else {
			log.Default().Printf("Log path should be either a nonexistent *.log file or a directory: %s", err)
		}
	}
	cmd := exec.Command(c.PythonRuntime, args...)
	cmd.Dir = c.ModelPath

	log.Default().Printf("Starting Python process: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Python process: %w", err)
	}

	// Try to connect to the socket with retries
	var conn net.Conn
	var err error
	for range 10 {
		conn, err = net.Dial("unix", socketPath)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		os.Remove(socketPath)
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}

	m := &SimpleModel{
		client: &ModelClient{
			socketPath: socketPath,
			conn:       conn,
			cmd:        cmd,
		},
		command: cmd,
	}

	return m, nil
}

func resolveLogPath(c *Config) (string, error) {
	info, err := os.Stat(c.LogPath)
	if err == nil && info.IsDir() {
		name := fmt.Sprintf("%d-peer_%s.log", time.Now().Unix(), c.Name)
		return filepath.Join(c.LogPath, name), nil
	} else if os.IsNotExist(err) && strings.HasSuffix(c.LogPath, ".log") {
		return c.LogPath, nil
	} else {
		return "", fmt.Errorf("unable to determine log file path: %w", err)
	}
}

// getRatio calculates the ratio of the model's own age as used by the Python model.
func getRatio(m *SimpleModel, weights Weights) float32 {
	ratio := float32(m.age) / (float32(m.age) + float32(weights.GetAge()))
	return ratio
}

// updateAge updates the model's age to the maximum of the current and the weights age.
func updateAge(m *SimpleModel, weights Weights) {
	tmp := max(m.age, weights.GetAge())
	m.age = tmp
}
