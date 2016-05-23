package web

import (
	"fmt"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/backend"
	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/oss"
	"github.com/containerops/dockyard/router"
	"github.com/containerops/dockyard/synch"
	"github.com/containerops/dockyard/utils/db"
	"github.com/containerops/dockyard/utils/setting"
)

func SetDockyardMacaron(m *macaron.Macaron) {
	if err := db.RegisterDriver(setting.DBDriver); err != nil {
		fmt.Printf("Register database driver error: %s\n", err.Error())
	} else {
		db.Drv.RegisterModel(new(models.Tag), new(models.Image), new(models.Repository))

		db.Drv.RegisterModel(new(dao.Organization), new(dao.User), new(dao.OrganizationUserMap),
			new(dao.RepositoryEx), new(dao.Team), new(dao.TeamRepositoryMap), new(dao.TeamUserMap))

		db.Drv.RegisterModel(new(models.Region), new(models.RegionTable))

		if err := db.Drv.InitDB(
			setting.DBDriver,
			setting.DBUser,
			setting.DBPasswd,
			setting.DBURI,
			setting.DBName,
			setting.DBDB); err != nil {
			fmt.Printf("Connect database error: %s\n", err.Error())
		}

		if err := dao.InitDAO(); err != nil {
			fmt.Printf("Init database access object error: %s\n", err.Error())
		}
	}

	if setting.Backend != "" {
		if err := backend.RegisterDriver(setting.Backend); err != nil {
			fmt.Printf("Register backend driver error: %s\n", err.Error())
		}
	}

	if err := middleware.Initfunc(); err != nil {
		fmt.Printf("Init middleware error: %s\n", err.Error())
	}

	//Setting Middleware
	middleware.SetMiddlewares(m)

	//Setting Router
	router.SetRouters(m)

	//Start Object Storage Service if sets in conf
	if strings.EqualFold(setting.OssSwitch, "enable") {
		ossobj := oss.Instance()
		ossobj.StartOSS()
	}

	//TODO:
	if setting.SynMode != "" {
		if err := synch.InitSynchron(); err != nil {
			fmt.Printf("Init synch error: %s\n", err.Error())
		}
	}
}
