package peer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vs-ude/btfl/internal/structs"
)

type Config struct {
	Name       string
	TrackerURL string
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
	return nil
}
