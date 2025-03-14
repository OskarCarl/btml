package tracker

import (
	"encoding/json"
	"time"

	"github.com/vs-ude/btml/internal/structs"
)

type Config struct {
	Tracker struct {
		PeerTimeout      time.Duration `toml:"peer_timeout"`
		MaintainInterval time.Duration `toml:"maintain_interval"`
		MaxPeers         int           `toml:"max_peers"`
		MaxReturnPeers   int           `toml:"max_return_peers"`
	} `toml:"tracker"`
	Peer struct {
		Dataset    string        `toml:"dataset"`
		UpdateFreq time.Duration `toml:"update_freq"`
	} `toml:"peer"`
	TelConf *structs.TelemetryConf `toml:"telemetry"`
}

func (c *Config) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}
