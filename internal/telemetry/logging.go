package telemetry

import "log/slog"

func (c *Client) ErrorLogging() {
	for err := range c.writeAPI.Errors() {
		slog.Warn("Sending telemetry failed", "error", err)
	}
}

func log(point string) {
	slog.Debug("Telemetry send", "point", point)
}
