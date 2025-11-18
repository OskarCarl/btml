package model

import (
	context "context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	grpc "google.golang.org/grpc"
)

// ModelClient implements the Model interface by communicating with a Python process
type ModelClient struct {
	socketPath          string
	conn                *grpc.ClientConn
	trainClient         TrainClient
	evalClient          EvalClient
	importWeightsClient ImportWeightsClient
	exportWeightsClient ExportWeightsClient
	cmd                 *exec.Cmd
}

func (c *ModelClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.cmd != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
		c.cmd.Wait()
	}
	return nil
}

func (c *ModelClient) Train() (*Metrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	res, err := c.trainClient.Train(ctx, &TrainRequest{})
	if err != nil {
		return nil, fmt.Errorf("train request failed: %w", err)
	}
	if !res.Success {
		return nil, fmt.Errorf("train request failed: %s", res.ErrorMessage)
	}
	return NewMetrics(-1, res.Loss, nil)
}

func (c *ModelClient) Eval(checkpointPath string) (*Metrics, error) {
	req := &EvalRequest{
		Path: checkpointPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	res, err := c.evalClient.Eval(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("eval request failed: %w", err)
	}
	if !res.Success {
		return nil, fmt.Errorf("eval request failed: %s", res.ErrorMessage)
	}
	return NewMetrics(res.Accuracy, res.Loss, res.Guesses)
}

func (c *ModelClient) Apply(weights *Weights, ratio float32) error {
	req := &ImportRequest{
		Weights:     weights.Get(),
		WeightRatio: ratio,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	res, err := c.importWeightsClient.ImportWeights(ctx, req)
	if err != nil {
		return fmt.Errorf("import weights request failed: %w", err)
	}
	if !res.Success {
		return fmt.Errorf("import weights request failed: %s", res.ErrorMessage)
	}
	return nil
}

func (c *ModelClient) GetWeights() (*Weights, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	res, err := c.exportWeightsClient.ExportWeights(ctx, &ExportRequest{})
	if err != nil {
		return nil, fmt.Errorf("export weights request failed: %w", err)
	}
	if !res.Success {
		return nil, fmt.Errorf("export weights request failed: %s", res.ErrorMessage)
	}
	return NewWeights(res.Weights, -1), nil
}
