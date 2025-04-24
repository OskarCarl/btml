package telemetry

import (
	"context"
	"fmt"

	influxdb3 "github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

type Client struct {
	client *influxdb3.Client
	ctx    context.Context
	cancel func()
	name   string
	tags   map[string]string
	run    string
}

func NewClient(conf *TelemetryConf, peerID string) (*Client, error) {
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     conf.URL,
		Token:    conf.Token,
		Database: conf.DB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create InfluxDB client: %w", err)
	}

	// Basic tags that will be added to all points
	tags := map[string]string{
		"peer_id": peerID,
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		client: client,
		ctx:    ctx,
		cancel: cancel,
		name:   peerID,
		tags:   tags,
		run:    conf.Suffix,
	}, nil
}

func (c *Client) Close() {
	c.client.Close()
	c.cancel()
}
