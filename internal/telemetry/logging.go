package telemetry

import "log/slog"

func log_w(err error) {
	slog.Warn("Sending telemetry failed", "error", err)
}

func log(point string) {
	slog.Debug("Telemetry send", "point", point)
}
