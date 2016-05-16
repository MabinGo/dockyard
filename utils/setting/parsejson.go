package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type NotificationsDesc struct {
	Name      string         `json:"name,omitempty"`
	Endpoints []EndpointDesc `json:"endpoints,omitempty"`
}

type EndpointDesc struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Headers   http.Header   `json:"headers"`
	Timeout   time.Duration `json:"timeout"`
	Threshold int           `json:"threshold"`
	Backoff   time.Duration `json:"backoff"`
	EventDB   string        `json:"eventdb"`
	Disabled  bool          `json:"disabled"`
}

type AuthorItem map[string]interface{}

type AuthorDesc map[string]AuthorItem

type JsonDesc struct {
	Notifications NotificationsDesc `json:"notifications,omitempty"`
	Authors       AuthorDesc        `json:"auth,omitempty"`
}

var JsonConf JsonDesc

func GetConfFromJSON(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	if err := json.Unmarshal(buf, &JsonConf); err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	return nil
}

func (auth AuthorDesc) Name() string {
	name := ""
	for key, _ := range auth {
		name = key
		break
	}
	return name
}
