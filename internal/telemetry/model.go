package telemetry

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func (c *Client) RecordTraining(loss float32, age int) {
	point := influxdb2.NewPoint(
		"model_training",
		c.tags,
		map[string]any{
			"loss": loss,
			"age":  age,
		},
		time.Now(),
	)

	log("model_training")
	c.writeAPI.WritePoint(point)
}

	point := influxdb2.NewPoint(
func (c *Client) RecordEvaluation(accuracy, loss float32, guesses map[int32]float32, age int) {
		"model_evaluation",
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
	c.writeAPI.WritePoint(point)
}

func (c *Client) RecordWeightApplication(localAge, remoteAge int) {
	point := influxdb2.NewPoint(
		"weight_application",
		c.tags,
		map[string]any{
			"local_age":  localAge,
			"remote_age": remoteAge,
		},
		time.Now(),
	)

	log("weight_application")
	c.writeAPI.WritePoint(point)
}
