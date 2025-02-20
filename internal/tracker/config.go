package tracker

import (
	"encoding/json"
	"time"
)

type Config struct {
	PeerTimeout       time.Duration `toml:"peer_timeout"`
	MaintainInterval  time.Duration `toml:"maintain_interval"`
	DatafilesBasePath string        `toml:"datafiles_base_path"`
	MetricURL         string        `toml:"metric_url"`
}

func (c *Config) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}
