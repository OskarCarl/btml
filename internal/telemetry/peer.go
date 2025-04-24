package telemetry

import (
	"fmt"
	"strings"
	"time"

	influxdb3 "github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

func (c *Client) RecordSend(age int, target string) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("peer_send_%s", c.run),
		c.tags,
		map[string]any{
			"age":    age,
			"source": c.name,
			"target": target,
		},
		time.Now(),
	)

	log("peer_send")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}

func (c *Client) RecordOnline(age int) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("peer_online_%s", c.run),
		c.tags,
		map[string]any{
			"id":  c.name,
			"age": age,
		},
		time.Now(),
	)

	log("peer_online")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}

func (c *Client) RecordActivePeers(peers []string) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("peer_active_%s", c.run),
		c.tags,
		map[string]any{
			"id":    c.name,
			"peers": strings.Join(peers, ","),
		},
		time.Now(),
	)

	log("peer_active")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}
