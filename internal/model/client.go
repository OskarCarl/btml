package model

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os/exec"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
)

// ModelClient implements the Model interface by communicating with a Python process
type ModelClient struct {
	socketPath string
	conn       net.Conn
	cmd        *exec.Cmd
}

func (c *ModelClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	if c.cmd != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
		c.cmd.Wait()
	}
	return nil
}

func (c *ModelClient) Train() (*Metrics, error) {
	req := &ModelRequest{
		Request: &ModelRequest_Train{
			Train: &TrainRequest{},
		},
	}

	err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("train request failed: %w", err)
	}

	res, err := c.readResponse(20)
	if err != nil {
		return nil, fmt.Errorf("train request failed: %w", err)
	}
	return NewMetrics(-1, res.Loss)
}

func (c *ModelClient) Eval(checkpointPath string) (*Metrics, error) {
	req := &ModelRequest{
		Request: &ModelRequest_Eval{
			Eval: &EvalRequest{
				Path: checkpointPath,
			},
		},
	}

	err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("eval request failed: %w", err)
	}

	res, err := c.readResponse(5)
	if err != nil {
		return nil, fmt.Errorf("eval request failed: %w", err)
	}
	return NewMetrics(res.Accuracy, res.Loss)
}

func (c *ModelClient) Apply(weights *Weights, ratio float32) error {
	req := &ModelRequest{
		Request: &ModelRequest_ImportWeights{
			ImportWeights: &ImportWeightsRequest{
				Weights:     weights.Get(),
				WeightRatio: ratio,
			},
		},
	}

	err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("import weights request failed: %w", err)
	}

	_, err = c.readResponse(5)
	if err != nil {
		return fmt.Errorf("import weights request failed: %w", err)
	}
	return nil
}

func (c *ModelClient) GetWeights() (*Weights, error) {
	req := &ModelRequest{
		Request: &ModelRequest_ExportWeights{
			ExportWeights: &ExportWeightsRequest{},
		},
	}

	err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("export weights request failed: %w", err)
	}

	res, err := c.readResponse(5)
	if err != nil {
		return nil, fmt.Errorf("export weights request failed: %w", err)
	}
	return NewWeights(res.Weights, -1), nil
}

func (c *ModelClient) sendRequest(req *ModelRequest) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write length-prefixed message
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(data)))
	if _, err := c.conn.Write(buf[:n]); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// Read response length
	ack, err := binary.ReadUvarint(newConnReader(c.conn))
	if err != nil || ack != 42 {
		return fmt.Errorf("failed to read ack: %d, %w", ack, err)
	}

	return nil
}

// readResponse waits for the response message for `timeout` seconds and returns it
func (c *ModelClient) readResponse(timeout int) (*ModelResponse, error) {
	packLen := make(chan uint64, 1)
	go func() {
		l, err := binary.ReadUvarint(newConnReader(c.conn))
		if err == nil {
			packLen <- l
		}
	}()
	select {
	case <-time.After(time.Second * time.Duration(timeout)):
		return nil, fmt.Errorf("timed out after %d seconds waiting for response", timeout)
	case l := <-packLen:
		// Read response data
		resData := make([]byte, l)
		if _, err := io.ReadFull(c.conn, resData); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Unmarshal res
		var res ModelResponse
		if err := proto.Unmarshal(resData, &res); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if !res.Success {
			return nil, fmt.Errorf("got error from Python: %s", res.ErrorMessage)
		}
		return &res, nil
	}
}

// connReader adapts a net.Conn to io.ByteReader for binary.ReadUvarint
type connReader struct {
	conn net.Conn
}

func newConnReader(conn net.Conn) *connReader {
	return &connReader{conn: conn}
}

func (r *connReader) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := r.conn.Read(buf)
	return buf[0], err
}
