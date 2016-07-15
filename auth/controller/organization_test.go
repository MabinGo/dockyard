package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
	"github.com/containerops/dockyard/utils/setting"
)

var openDBFlag bool = false

func init() {
	setting.Authn = "authn_db"
}

func openDB(t *testing.T) {
	if openDBFlag {
		return
	}
	if err := db.RegisterDriver("mysql"); err != nil {
		t.Fatal(err)
	} else {
		db.Drv.RegisterModel(new(dao.Organization), new(dao.User), new(dao.OrganizationUserMap),
			new(dao.RepositoryEx), new(dao.Team), new(dao.TeamRepositoryMap), new(dao.TeamUserMap))
		err := db.Drv.InitDB("mysql", "root", "root", "127.0.0.1:3306", "dockyard", 0)
		if err != nil {
			t.Fatal(err)
		}
	}

	//4. create fk and root user
	if err := dao.InitDAO(); err != nil {
		t.Fatal(err)
	}

	//5. open authz
	//if err := authz.AuthorizerOpen(); err != nil {
	//	t.Fatal(err)
	//}

	openDBFlag = true
}

func Test_SetupEnv(t *testing.T) {
	openDB(t)

	//start authn
	var err error
	authn.Authn, err = authn.NewAuthenticator()
	if err != nil {
		t.Fatal(fmt.Errorf("New Authenticator error: %s\n", err.Error()))
	}
}

func CreateOrganizationTest(t *testing.T, org *dao.Organization, userName, passWord string) {
	b, err := json.Marshal(org)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", "127.0.0.1:8080\\organization", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, passWord)

	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := CreateOrganization(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_CreateOrganization(t *testing.T) {
	//1. create user
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		RealName: "liugenping",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Error(err)
	}

	//2. create organization
	org := &dao.Organization{
		Name: "huawei",
		//Email           string
		//Comment         string
		//URL             string
		//Location        string
		MemberPrivilege: dao.WRITE,
	}
	CreateOrganizationTest(t, org, user.Name, "liugenping")

	//3. query organization
	org1 := &dao.Organization{Name: "huawei"}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if org1.MemberPrivilege != org.MemberPrivilege {
			t.Error("org's save is not same with get")
		}
	}

	//4. del user and organization
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}

func UpdateOrganizationTest(t *testing.T, org *dao.Organization, userName, passWord string) {
	b, err := json.Marshal(org)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("PUT", "127.0.0.1:8080\\organization\\update", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, passWord)

	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := UpdateOrganization(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_UpdateOrganization(t *testing.T) {

	//1. create user
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		RealName: "liugenping",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Error(err)
	}

	//2. create organization
	org := &dao.Organization{
		Name:            "huaweicn",
		Email:           "org@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	CreateOrganizationTest(t, org, user.Name, "liugenping")

	//4. update organization
	org.Comment = "Comment update"
	org.MemberPrivilege = dao.READ
	UpdateOrganizationTest(t, org, user.Name, "liugenping")

	//5. query update result
	org2 := &dao.Organization{Name: org.Name}
	if exist, err := org2.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if (org2.MemberPrivilege != dao.READ) || (org2.Comment != org.Comment) {
			t.Error("update organization info failed")
		}
	}

	//6. del user and organization
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}
