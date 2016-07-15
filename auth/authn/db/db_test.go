package db

import (
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
)

var openDBFlag bool = false

func openDB(t *testing.T) {

	if openDBFlag {
		return
	}
	if err := db.RegisterDriver("mysql"); err != nil {
		t.Error(err)
	} else {
		db.Drv.RegisterModel(new(dao.User), new(dao.Organization))
		err := db.Drv.InitDB("mysql", "root", "root", "127.0.0.1:3306", "dockyard", 0)
		if err != nil {
			t.Error(err)
		}
	}
	openDBFlag = true
}

func Test_Authenticate(t *testing.T) {
	openDB(t)
	u := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "111111111",
		//RealName string `orm:"size(100)";null`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   1,
	}
	pwd := u.Password
	if err := u.Save(); err != nil {
		t.Error(err)
	}

	auth := DBAuthn{}
	if _, err := (&auth).Authenticate(u.Name, pwd); err != nil {
		t.Error(err)
	}

	if err := u.Delete(); err != nil {
		t.Error(err)
	}
}
