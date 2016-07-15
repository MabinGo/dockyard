package huaweiw3

import (
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
)

func Test_SetEnv(t *testing.T) {
	//1.open db
	if err := db.RegisterDriver("mysql"); err != nil {
		t.Fatal(err)
	} else {
		db.Drv.RegisterModel(new(dao.User), new(dao.Organization))
		err := db.Drv.InitDB("mysql", "root", "root", "127.0.0.1:3306", "dockyard", 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	//2.save user to db
	u := &dao.User{
		Name:     "l00257029",
		Email:    "liugenping@huawei.com",
		Password: "root",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := u.Save(); err != nil {
		t.Fatal(err)
	}
}

func Test_Authenticate_TLSSkipVerify(t *testing.T) {
	conf := &HuaweiW3AuthnConfig{

		Addr: "https://login.huawei.com/login/rest/token",
		InsecureTLSSkipVerify: true,
		//CertFile:              "/home/liugenping/gopath/src/github.com/containerops/dockyard/cert/huaweiw3/verisign/verisign.cer",
	}

	if h, err := NewHuaweiW3Authn(conf); err != nil {
		t.Error(err)
	} else {
		if _, err := h.Authenticate("l00257029", "fjeifjoef!"); err != nil {
			t.Error(err)
		}
	}
}

func Test_Authenticate_TLSVerify(t *testing.T) {
	//setting.Cert = "/home/liugenping/gopath/src/github.com/containerops/dockyard/cert/huaweiw3/verisign/verisign-ca1.cer"
	//setting.Cert = "/home/liugenping/gopath/src/github.com/containerops/dockyard/cert/containerops/containerops.crt"
	conf := &HuaweiW3AuthnConfig{
		Addr: "https://login.huawei.com/login/rest/token",
		InsecureTLSSkipVerify: false,
		CertFile:              "/home/liugenping/gopath/src/github.com/containerops/dockyard/cert/huaweiw3/verisign/verisign.cer",
	}

	if h, err := NewHuaweiW3Authn(conf); err != nil {
		t.Error(err)
	} else {
		if _, err := h.Authenticate("l00257029", "fjeifjoef!"); err != nil {
			t.Error(err)
		}
	}
}

func Test_Authenticate_ErrorCert(t *testing.T) {
	conf := &HuaweiW3AuthnConfig{
		Addr: "https://login.huawei.com/login/rest/token",
		InsecureTLSSkipVerify: false,
		CertFile:              "/home/liugenping/gopath/src/github.com/containerops/dockyard/cert/containerops/containerops.crt",
	}

	if h, err := NewHuaweiW3Authn(conf); err != nil {
		t.Error(err)
	} else {
		if _, err := h.Authenticate("l00257029", "fjeifjoef!"); err != nil {
			t.Log(err)
		} else {
			t.Error("not pass for error cert")
		}
	}
}

func Test_CleanEnv(t *testing.T) {
	u := &dao.User{
		Name: "l00257029",
	}
	if err := u.Delete(); err != nil {
		t.Error(err)
	}
}
