package model

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Model represents a model instance. All actions are executed in series.
type Model struct {
	client                *ModelClient
	age                   int
	modelModifiedCallback func(*Weights)
	sync.Mutex
}

// Shutdown closes the model client and logs a message. It ignores the lock.
func (m *Model) Shutdown() {
	m.client.Close()
	log.Default().Println("Model stopped")
}

// Eval evaluates the model and logs the results. It blocks until other
// operations are completed.
func (m *Model) Eval() error {
	m.Lock()
	defer m.Unlock()
	met, err := m.client.Eval()
	if err != nil {
		return fmt.Errorf("failed to evaluate model: %w", err)
	}
	log.Default().Printf("Evaluated model; acc: %f, loss: %f", met.acc, met.loss)
	return nil
}

// Train trains the model and logs the results. It blocks until other
// operations are completed.
func (m *Model) Train() error {
	m.Lock()
	defer m.Unlock()
	met, err := m.client.Train()
	if err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}
	m.age++
	log.Default().Printf("Trained model to age %d; loss: %f", m.age, met.loss)
	m.executeCallback()
	return nil
}

// Apply applies the given weights to the model, does a short training run, and
// logs the results. It blocks until other operations are completed.
func (m *Model) Apply(weights *Weights) error {
	m.Lock()
	defer m.Unlock()
	ratio := getRatio(m, weights)
	if err := m.client.Apply(weights, ratio); err != nil {
		return fmt.Errorf("failed to apply weights to model: %w", err)
	}
	met, err := m.client.Train()
	if err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}
	log.Default().Printf("Applied weights to model, loss: %f", met.loss)
	updateAge(m, weights)
	m.executeCallback()
	return nil
}

// GetWeights fetches the weights from the model and returns them. It blocks
// until other operations are completed.
func (m *Model) GetWeights() (*Weights, error) {
	m.Lock()
	defer m.Unlock()
	return m.getWeights()
}

// getWeights assumes that the model is locked.
func (m *Model) getWeights() (*Weights, error) {
	w, err := m.client.GetWeights()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weights from model: %w", err)
	}
	log.Default().Print("Got weights from model")
	w.setAge(m.age)
	return w, nil
}

// executeCallback runs the callback function if it is set. It uses getWeights
// so it assumes that the model is locked.
func (m *Model) executeCallback() {
	if m.modelModifiedCallback != nil {
		w, err := m.getWeights()
		if err != nil {
			log.Default().Printf("Failed to get weights for callback: %s", err)
			return
		}
		m.modelModifiedCallback(w)
	}
}

// NewModel creates a new Model instance by starting the Python process
// and establishing a connection to it
func NewModel(c *Config) (*Model, error) {
	// Create a random socket path in /tmp
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("btml-model-%d.sock", time.Now().Unix()))

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
	return &Model{
		client: &ModelClient{
			cmd:        cmd,
			socketPath: socketPath,
		},
		age:                   1,
		modelModifiedCallback: nil,
	}, nil
}

func (m *Model) SetCallback(callback func(*Weights)) {
	m.modelModifiedCallback = callback
}

func (m *Model) Start() error {
	log.Default().Printf("Starting Python process: %s (cwd: %s)", m.client.cmd.String(), m.client.cmd.Dir)
	if err := m.client.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python process: %w", err)
	}

	// Try to connect to the socket with retries
	var conn net.Conn
	var err error
	for i := range 10 {
		conn, err = net.Dial("unix", m.client.socketPath)
		if err == nil {
			break
		}
		if i < 4 && i > 1 {
			log.Default().Printf("No response from model, retrying %d/5 ...", i+2)
		}
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		m.client.cmd.Process.Kill()
		m.client.cmd.Wait()
		os.Remove(m.client.socketPath)
		return fmt.Errorf("failed to connect to socket: %w", err)
	}

	m.client.conn = conn

	return nil
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
func getRatio(m *Model, weights *Weights) float32 {
	ratio := float32(m.age) / (float32(m.age) + float32(weights.GetAge()))
	return ratio
}

// updateAge updates the model's age to the maximum of the current and the weights age.
func updateAge(m *Model, weights *Weights) {
	tmp := max(m.age, weights.GetAge())
	m.age = tmp
}
