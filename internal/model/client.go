package model

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os/exec"

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
	return nil
}

func (c *ModelClient) Train() error {
	req := &ModelRequest{
		Request: &ModelRequest_Train{
			Train: &TrainRequest{},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("train request failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("train failed: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *ModelClient) Eval() error {
	req := &ModelRequest{
		Request: &ModelRequest_Eval{
			Eval: &EvalRequest{},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("eval request failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("eval failed: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *ModelClient) Apply(weights Weights) error {
	req := &ModelRequest{
		Request: &ModelRequest_ImportWeights{
			ImportWeights: &ImportWeightsRequest{
				Weights:     weights.Get(),
				WeightRatio: 1.0, // TODO: calculate this
			},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("import weights request failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("import weights failed: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *ModelClient) GetWeights() (Weights, error) {
	req := &ModelRequest{
		Request: &ModelRequest_ExportWeights{
			ExportWeights: &ExportWeightsRequest{},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("export weights request failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("export weights failed: %s", resp.ErrorMessage)
	}

	return NewSimpleWeights(resp.Weights)
}

func (c *ModelClient) sendRequest(req *ModelRequest) (*ModelResponse, error) {
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write length-prefixed message
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(len(data)))
	if _, err := c.conn.Write(buf[:n]); err != nil {
		return nil, fmt.Errorf("failed to write message length: %w", err)
	}
	if _, err := c.conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	// Read response length
	respLen, err := binary.ReadUvarint(newConnReader(c.conn))
	if err != nil {
		return nil, fmt.Errorf("failed to read response length: %w", err)
	}

	// Read response data
	respData := make([]byte, respLen)
	if _, err := io.ReadFull(c.conn, respData); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal response
	var resp ModelResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
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
