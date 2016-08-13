package docker

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	"github.com/astaxie/beego/config"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	Domains      string
	UserName     string
	DockerBinary string
	user         *User
)

func TestGetDockerConf(t *testing.T) {
	path := "../testsuite.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}

	if domains := conf.String("test::domains"); domains != "" {
		Domains = domains
	} else {
		t.Errorf("Read %s error: nil", domains)
	}

	if username := conf.String("test::username"); username != "" {
		UserName = username
	} else {
		t.Errorf("Read %s error: nil", username)
	}

	if client := conf.String("test::client"); client != "" {
		DockerBinary = client
	} else {
		t.Errorf("Read %s error: nil", client)
	}
}

type User struct {
	Name     string
	Email    string
	Password string
	RealName string
	Comment  string
}

func TestLogin(t *testing.T) {
	if err := setting.SetConfig("../../conf/containerops.conf"); err != nil {
		t.Error(err)
	}

	if setting.Authmode == "token" {
		user = &User{
			Name:     UserName,
			Email:    "root@huawei.com",
			Password: "root",
			RealName: "root",
			Comment:  "commnet",
		}
		signUp(user, t)
		cmd := exec.Command("sudo", DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Docker login faild: [Error]%v", err)
		}
	}
}

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}

func signUp(user *User, t *testing.T) {
	b, err := json.Marshal(user)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/user/signup", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}
