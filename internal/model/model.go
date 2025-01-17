package model

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type Model interface {
	Eval() error
	Train() error
	Apply(Weights) error
	GetWeights() (Weights, error)
	Close() error
}

type SimpleModel struct {
	client  *ModelClient
	command *exec.Cmd
}

func (m *SimpleModel) Close() error {
	m.client.Close()
	m.command.Process.Signal(syscall.SIGTERM)
	return nil
}

func (m *SimpleModel) Eval() error {
	if err := m.client.Eval(); err != nil {
		return fmt.Errorf("failed to evaluate model: %w", err)
	}
	return m.client.Eval()
}

func (m *SimpleModel) Train() error {
	if err := m.client.Train(); err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}
	return nil
}

func (m *SimpleModel) Apply(weights Weights) error {
	if err := m.client.Apply(weights); err != nil {
		return fmt.Errorf("failed to apply weigths to model: %w", err)
	}
	return nil
}

func (m *SimpleModel) GetWeights() (Weights, error) {
	w, err := m.client.GetWeights()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weights from model: %w", err)
	}
	return w, nil
}

// NewModel creates a new Model instance by starting the Python process
// and establishing a connection to it
func NewSimpleModel(workdir, runtimePath, trainPath, testPath, logOutput string) (Model, error) {
	// Create a random socket path in /tmp
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("btfl-model-%d.sock", time.Now().UnixNano()))

	// Start the Python process
	args := []string{
		"main.py",
		"--train-data", trainPath,
		"--test-data", testPath,
		"--socket", socketPath,
	}
	if logOutput != "" {
		args = append(args, "--log-file", logOutput)
	}
	cmd := exec.Command(runtimePath, args...)
	cmd.Dir = workdir

	log.Default().Printf("Starting Python process: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Python process: %w", err)
	}

	// Try to connect to the socket with retries
	var conn net.Conn
	var err error
	for i := 0; i < 10; i++ {
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
