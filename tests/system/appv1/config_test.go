package appv1

import (
	"testing"

	"github.com/astaxie/beego/config"
)

var (
	Domains           string
	UserName          string
	DockerBinary      string
	DockyardURL       string
	DockyardURLSearch string
	ListenMode        string
)

func TestGetDockerConf(t *testing.T) {
	path := "../testsuite.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}
	if listenmode := conf.String("test_appv1::listenmode"); listenmode != "" {
		ListenMode = listenmode
	} else {
		t.Errorf("Read %s error: nil", listenmode)
	}

	if domains := conf.String("test::domains"); domains != "" {
		Domains = domains
		DockyardURL = ListenMode + "://" + domains + "/" + "app" + "/" + "v1" + "/"
		DockyardURLSearch = ListenMode + "://" + domains + "/" + "app" + "/" + "v1" + "/" + "search?key="

	} else {
		t.Errorf("Read %s error: nil", domains)
	}
}
