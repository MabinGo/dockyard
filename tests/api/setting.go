package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego/config"
)

var (
	Domains               string
	UserName              string
	DockerBinary          string
	ListenMode            string
	DockerFrom            string
	DockyardURI           string
	HTTPReconnectionCount int
	AuthEnable            bool
	AuthKey               string
	AuthValue             string
	Token                 = ""
	dockyardHealth        bool
	openAuth              bool
)

func SetConfig(path string) error {
	if Domains != "" {
		return nil
	}
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		return fmt.Errorf("Read %s error: %v", path, err)
	}
	if listenMode := conf.String("test_appv1::listenmode"); listenMode != "" {
		ListenMode = listenMode
	} else {
		return errors.New("ListenMode config value is null")
	}

	if domains := conf.String("test::domains"); domains != "" {
		Domains = domains
		DockyardURI = fmt.Sprintf("%s://%s", ListenMode, domains)
	} else {
		return errors.New("Domains config value is null")
	}

	if username := conf.String("test::username"); username != "" {
		UserName = username
	} else {
		return errors.New("UserName config value is null")
	}

	if client := conf.String("test::client"); client != "" {
		DockerBinary = client
	} else {
		return errors.New("DockerBinary config value is null")
	}

	if dockerFrom := conf.String("test::dockerfrom"); dockerFrom != "" {
		DockerFrom = dockerFrom
	} else {
		DockerFrom = "busybox:latest"
	}

	if reconnectionCount, err := strconv.Atoi(conf.String("test::HTTPReconnectionCount")); err == nil && reconnectionCount > 0 {
		HTTPReconnectionCount = reconnectionCount
	} else {
		HTTPReconnectionCount = 1
	}

	if authEnable := conf.String("auth::enable"); strings.ToLower(authEnable) == "true" {
		AuthEnable = true
		if authKey := conf.String("auth::key"); authKey != "" {
			AuthKey = authKey
		} else {
			return errors.New("Auth:key config is null when enable is true")
		}
		if authValue := conf.String("auth::value"); authValue != "" {
			AuthValue = authValue
		} else {
			return errors.New("Auth:value config is null when enable is true")
		}
	}
	return nil
}
