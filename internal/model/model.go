package model

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vs-ude/btml/internal/telemetry"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Model represents a model instance. All actions are executed in series.
type Model struct {
	checkpointBase        string
	client                *ModelClient
	age                   int
	lastEval              int
	modelModifiedCallback func(*Weights)
	telemetry             *telemetry.Client
	sync.Mutex
}

// Shutdown closes the model client and logs a message. It ignores the lock.
func (m *Model) Shutdown() {
	m.client.Close()
	slog.Info("Model stopped")
}

// Eval evaluates the model and logs the results. It blocks until other
// operations are completed.
func (m *Model) Eval() error {
	checkpointPath := fmt.Sprintf("%s/%d", m.checkpointBase, m.age)
	m.Lock()
	defer m.Unlock()
	met, err := m.client.Eval(checkpointPath)
	if err != nil {
		return fmt.Errorf("failed to evaluate model: %w", err)
	}
	slog.Info("Evaluated model", "accuracy", met.acc, "loss", met.loss)
	if m.telemetry != nil {
		go m.telemetry.RecordEvaluation(met.acc, met.loss, met.guesses, m.age)
	}
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
	slog.Info("Trained model", "age", m.age, "loss", met.loss)
	if m.telemetry != nil {
		go m.telemetry.RecordTraining(met.loss, m.age)
	}
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
	if m.telemetry != nil {
		go m.telemetry.RecordWeightApplication(m.age, weights.GetAge())
	}
	met, err := m.client.Train()
	if err != nil {
		return fmt.Errorf("failed to train model: %w", err)
	}
	m.age = max(m.age, weights.GetAge()) + 1
	if m.telemetry != nil {
		go m.telemetry.RecordTraining(met.loss, m.age)
	}
	slog.Info("Applied weights to model", "age", m.age, "loss", met.loss)
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
	slog.Debug("Got weights from model")
	w.SetAge(m.age)
	return w, nil
}

// executeCallback runs the callback function if it is set. It uses getWeights
// so it assumes that the model is locked.
func (m *Model) executeCallback() {
	if m.modelModifiedCallback != nil {
		w, err := m.getWeights()
		if err != nil {
			slog.Error("Failed to get weights for callback", "error", err)
			return
		}
		m.modelModifiedCallback(w)
	}
}

// NewModel creates a new Model instance by starting the Python process
// and establishing a connection to it
func NewModel(c *Config, telemetry *telemetry.Client) (*Model, error) {
	// Create a random socket path in /tmp
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("btml-model-%d.sock", time.Now().Unix()))

	args := append(c.ModelArgs,
		"--train-data", c.GetTrainDataPath(),
		"--test-data", c.GetTestDataPath(),
		"--socket", socketPath,
	)
	if c.LogPath != "" {
		if p, err := resolveLogPath(c); err == nil {
			args = append(args, "--log-file", p)
		} else {
			slog.Warn("Invalid log path configuration. Log path should be either a nonexistent *.log file or a directory.", "error", err)
		}
	}
	cmd := exec.Command(c.PythonRuntime, args...)
	stdout, _ := cmd.StderrPipe()
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			slog.Error("Model error", "text", scanner.Text())
		}
	}()
	return &Model{
		client: &ModelClient{
			cmd:        cmd,
			socketPath: socketPath,
		},
		age:                   1,
		checkpointBase:        c.GetCheckpointPath(),
		modelModifiedCallback: nil,
		telemetry:             telemetry,
	}, nil
}

func (m *Model) SetCallback(callback func(*Weights)) {
	m.modelModifiedCallback = callback
}

func (m *Model) Start() error {
	slog.Info("Starting Python process", "command", m.client.cmd.String(), "cwd", m.client.cmd.Dir)
	if err := m.client.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python process: %w", err)
	}

	// Try to connect to the socket with retries
	var conn *grpc.ClientConn
	var err error
	conn, err = grpc.NewClient("unix://"+m.client.socketPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}
	maxAttempts := 10
	for i := range maxAttempts {
		conn.Connect()
		state := conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			break
		}
		if i == maxAttempts {
			m.client.cmd.Process.Kill()
			m.client.cmd.Wait()
			os.Remove(m.client.socketPath)
			return fmt.Errorf("failed to connect to socket: %w", err)
		}
		if i > 3 {
			slog.Debug("No response from model (yet)", "attempt", i+1, "max_attempts", maxAttempts)
		}
		time.Sleep(time.Second * 2)
	}

	m.client.trainClient = NewTrainClient(conn)
	m.client.evalClient = NewEvalClient(conn)
	m.client.exportWeightsClient = NewExportWeightsClient(conn)
	m.client.importWeightsClient = NewImportWeightsClient(conn)

	slog.Info("Model process is set up and running")

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
	ratio := float32(weights.GetAge()) / (float32(m.age) + float32(weights.GetAge()))
	return ratio
}

func (m *Model) GetAge() int {
	return m.age
}

func (m *Model) EvalLoop() {
	timer := time.NewTimer(time.Second * 5)
	for {
		<-timer.C
		if m.age <= m.lastEval {
			timer.Reset(time.Second * 30)
			continue
		}
		m.Eval()
		timer.Reset(time.Second * 30)
	}
}
