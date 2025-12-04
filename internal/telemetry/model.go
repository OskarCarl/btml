package telemetry

import (
	"fmt"
	"time"

	influxdb3 "github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

func (c *Client) RecordTraining(loss float32, age int) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("model_training_%s", c.run),
		c.tags,
		map[string]any{
			"loss": loss,
			"age":  age,
		},
		time.Now(),
	)

	log("model_training")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}

func (c *Client) RecordEvaluation(accuracy, loss float32, guesses map[int32]float32, age int) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("model_evaluation_%s", c.run),
		c.tags,
		map[string]any{
			"accuracy": accuracy,
			"loss":     loss,
			"guesses":  guesses,
			"age":      age,
		},
		time.Now(),
	)

	log("model_evaluation")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}

func (c *Client) RecordWeightApplication(localAge, remoteAge int, change float32) {
	point := influxdb3.NewPoint(
		fmt.Sprintf("weight_application_%s", c.run),
		c.tags,
		map[string]any{
			"local_age":  localAge,
			"remote_age": remoteAge,
			"change":     change,
		},
		time.Now(),
	)

	log("weight_application")
	err := c.client.WritePoints(c.ctx, []*influxdb3.Point{point})
	if err != nil {
		log_w(err)
	}
}
