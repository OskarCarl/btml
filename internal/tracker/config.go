package tracker

import (
	"encoding/json"
	"time"
)

type Config struct {
	Tracker struct {
		PeerTimeout      time.Duration `toml:"peer_timeout"`
		MaintainInterval time.Duration `toml:"maintain_interval"`
		MaxPeers         int           `toml:"max_peers"`
	} `toml:"tracker"`
	Peer struct {
		MetricURL  string        `toml:"metric_url"`
		Dataset    string        `toml:"dataset"`
		UpdateFreq time.Duration `toml:"update_freq"`
	} `toml:"peer"`
}

func (c *Config) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}
