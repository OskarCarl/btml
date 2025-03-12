package structs

import (
	"time"
)

type Metrics struct {
	URL   string `toml:"url"`
	DB    string `toml:"db"`
	Token string `toml:"token"`
}

type WhoAmI struct {
	Id         int
	Dataset    string
	UpdateFreq time.Duration
	ExtIp      string
	Metrics    Metrics
}
