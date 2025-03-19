package telemetry

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func (c *Client) RecordSend(age int, target string) {
	point := influxdb2.NewPoint(
		"peer_send",
		c.tags,
		map[string]any{
			"age":    age,
			"source": c.name,
			"target": target,
		},
		time.Now(),
	)

	log("peer_send")
	c.writeAPI.WritePoint(point)
}

func (c *Client) RecordOnline(age int) {
	point := influxdb2.NewPoint(
		"peer_online",
		c.tags,
		map[string]any{
			"id":  c.name,
			"age": age,
		},
		time.Now(),
	)

	log("peer_online")
	c.writeAPI.WritePoint(point)
}
