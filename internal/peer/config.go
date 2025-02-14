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
	Dataset    string
	Basepath   string
}

func GetTrainPath(c *Config) string {
	return fmt.Sprintf("%s/%s_train_split_%s.pt", c.Basepath, c.Dataset, c.Name)
}

func GetTestPath(c *Config) string {
	return fmt.Sprintf("%s/%s_test_split_%s.pt", c.Basepath, c.Dataset, c.Name)
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
	c.Dataset = whoami.Dataset
	return nil
}
