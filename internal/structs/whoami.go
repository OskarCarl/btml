package structs

import (
	"time"

	"github.com/vs-ude/btml/internal/telemetry"
)

type WhoAmI struct {
	Id         int
	Dataset    string
	UpdateFreq time.Duration
	ExtIp      string
	Telemetry  telemetry.TelemetryConf
}
