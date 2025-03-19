package telemetry

import (
	"context"
	"fmt"

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
		"peer_id": peerID,
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
