package authtest

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/astaxie/beego/config"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	Domains      string
	UserName     string
	DockerBinary string
	LdapUserName string
	LdapPassword string
	DockyardPath string
)

func GetConfig() {
	gopath := os.Getenv("GOPATH")
	if len(gopath) <= 0 {
		panic("should set GOPATH")
	}
	DockyardPath = gopath + "/src/github.com/containerops/dockyard"
	if err := setting.SetConfig(DockyardPath + "/conf/containerops.conf"); err != nil {
		panic(err)
	}

	path := DockyardPath + "/test/testsuite.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		panic(fmt.Sprintf("Read %s error: %v", path, err.Error()))
	}

	if domains := conf.String("test::domains"); domains != "" {
		Domains = domains
	} else {
		panic(fmt.Sprintf("Read %s error: nil", domains))
	}

	if username := conf.String("test::username"); username != "" {
		UserName = username
	} else {
		panic(fmt.Sprintf("Read %s error: nil", username))
	}

	if client := conf.String("test::client"); client != "" {
		DockerBinary = client
	} else {
		panic(fmt.Sprintf("Read %s error: nil", client))
	}

	if ldapusername := conf.String("test::ldapusername"); ldapusername != "" {
		LdapUserName = ldapusername
	} else {
		panic(fmt.Sprintf("Read %s error: nil", ldapusername))
	}

	if ldappassword := conf.String("test::ldappassword"); ldappassword != "" {
		LdapPassword = ldappassword
	} else {
		panic(fmt.Sprintf("Read %s error: nil", ldappassword))
	}
}

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}
