package telemetry

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/vs-ude/btml/internal/structs"
)

type Client struct {
	client   influxdb2.Client
	name     string
	writeAPI api.WriteAPI
	tags     map[string]string
}

func NewClient(conf *structs.TelemetryConf, peerID string) (*Client, error) {
	client := influxdb2.NewClient(conf.URL, conf.Token)
	writeAPI := client.WriteAPI(conf.Org, conf.Bucket)

	// Basic tags that will be added to all points
	tags := map[string]string{
		// "peer_id": peerID,
	}

	online, err := client.Ping(context.Background())
	if !online || err != nil {
		return nil, fmt.Errorf("failed to ping InfluxDB: %w", err)
	}

	return &Client{
		client:   client,
		name:     peerID,
		writeAPI: writeAPI,
		tags:     tags,
	}, nil
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) RecordTraining(loss float32, age int) {
	point := influxdb2.NewPoint(
		"model_training",
		c.tags,
		map[string]any{
			"peer": c.name,
			"loss": loss,
			"age":  age,
		},
		time.Now(),
	)

	c.writeAPI.WritePoint(point)
}

func (c *Client) RecordEvaluation(accuracy, loss float32, age int) {
	point := influxdb2.NewPoint(
		"model_evaluation",
		c.tags,
		map[string]any{
			"peer":     c.name,
			"accuracy": accuracy,
			"loss":     loss,
			"age":      age,
		},
		time.Now(),
	)

	c.writeAPI.WritePoint(point)
}

func (c *Client) RecordWeightApplication(localAge, remoteAge int) {
	point := influxdb2.NewPoint(
		"weight_application",
		c.tags,
		map[string]any{
			"peer":       c.name,
			"local_age":  localAge,
			"remote_age": remoteAge,
		},
		time.Now(),
	)

	c.writeAPI.WritePoint(point)
}
