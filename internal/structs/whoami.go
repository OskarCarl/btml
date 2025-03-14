package structs

import (
	"time"
)

type TelemetryConf struct {
	URL    string `toml:"url"`
	Org    string `toml:"org"`
	Bucket string `toml:"bucket"`
	Token  string `toml:"token"`
}

type WhoAmI struct {
	Id         int
	Dataset    string
	UpdateFreq time.Duration
	ExtIp      string
	Telemetry  TelemetryConf
}
