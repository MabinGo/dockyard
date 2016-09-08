/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/db"
	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/web"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "start dockyard web service",
	Description: "dockyard is the module of handler docker and rkt image.",
	Action:      runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "web service listen ip, default is 0.0.0.0; if listen with Unix Socket, the value is sock file path.",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 80,
			Usage: "web service listen at port 80; if run with https will be 443.",
		},
	},
}

func runWeb(c *cli.Context) {
	middleware.InitLogger()
	if err := initDockyardDB(); err != nil {
		log.Errorf("Init Dockyard database error: %v", err.Error())
		os.Exit(1)
	}
	m := macaron.New()

	//Set Macaron Web Middleware And Routers
	web.SetDockyardMacaron(m)

	switch setting.ListenMode {
	case "http":
		listenaddr := fmt.Sprintf("%s:%d", c.String("address"), c.Int("port"))
		if err := http.ListenAndServe(listenaddr, m); err != nil {
			log.Errorf("Start Dockyard http service error: %v", err.Error())
			os.Exit(1)
		}
		break
	case "https":
		listenaddr := fmt.Sprintf("%s:443", c.String("address"))
		server := &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
		if err := server.ListenAndServeTLS(setting.HttpsCertFile, setting.HttpsKeyFile); err != nil {
			log.Errorf("Start Dockyard https service error: %v", err.Error())
			os.Exit(1)
		}
		break
	case "unix":
		listenaddr := fmt.Sprintf("%s", c.String("address"))
		if utils.IsFileExist(listenaddr) {
			os.Remove(listenaddr)
		}

		if listener, err := net.Listen("unix", listenaddr); err != nil {
			log.Errorf("Start Dockyard unix socket error: %v", err.Error())
			os.Exit(1)
		} else {
			server := &http.Server{Handler: m}
			if err := server.Serve(listener); err != nil {
				log.Errorf("Start Dockyard unix socket error: %v", err.Error())
				os.Exit(1)
			}
		}
		break
	default:
		break
	}
}

func initDockyardDB() error {
	if err := db.InitDB(setting.DatabaseDriver, setting.DatabaseUser, setting.DatabasePasswd,
		setting.DatabaseURI, setting.DatabaseName); err != nil {
		return err
	}
	if err := db.Instance.RegisterModel(new(models.AppV1), new(models.ArtifactV1), new(models.Session)); err != nil {
		return err
	}
	if err := db.Instance.RegisterModel(new(models.DockerV2), new(models.DockerImageV2),
		new(models.DockerTagV2)); err != nil {
		return err
	}
	if err := new(models.AppV1).AddUniqueIndex(); err != nil {
		return err
	}
	if err := new(models.ArtifactV1).AddUniqueIndex(); err != nil {
		return err
	}
	if err := new(models.Session).AddUniqueIndex(); err != nil {
		return err
	}
	if err := new(models.DockerV2).AddUniqueIndex(); err != nil {
		return err
	}
	if err := new(models.DockerImageV2).AddUniqueIndex(); err != nil {
		return err
	}
	if err := new(models.DockerTagV2).AddUniqueIndex(); err != nil {
		return err
	}

	return nil
}
