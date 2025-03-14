package telemetry

import "log/slog"

func (c *Client) ErrorLogging() {
	for err := range c.writeAPI.Errors() {
		slog.Warn("Sending telemetry failed", "error", err)
	}
}
