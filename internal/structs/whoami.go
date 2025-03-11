package structs

import (
	"time"
)

type WhoAmI struct {
	Id         int
	Dataset    string
	UpdateFreq time.Duration
	ExtIp      string
}
