package peer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/structs"
)

type Config struct {
	Name       string
	TrackerURL string
	UpdateFreq time.Duration
	ModelConf  *model.Config
}

func Autoconf(c *Config) error {
	resp, err := http.Get(c.TrackerURL + "/whoami")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := getResponseBody(resp)
	if err != nil {
		return err
	}
	whoami := new(structs.WhoAmI)
	err = json.Unmarshal(*body, whoami)
	if err != nil {
		return fmt.Errorf("unable to parse whoami response body data from tracker\n%w", err)
	}

	c.Name = strconv.Itoa(whoami.Id)
	c.UpdateFreq = whoami.UpdateFreq
	c.ModelConf.Dataset = whoami.Dataset
	c.ModelConf.Name = c.Name

	return nil
}
