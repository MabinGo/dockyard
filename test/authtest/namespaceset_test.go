package authtest

import (
	"testing"

	"github.com/astaxie/beego/config"

	"github.com/containerops/dockyard/auth/dao"
)

var user *dao.User
var org *dao.Organization

func Test_NamespaceSetInit(t *testing.T) {
	GetConfig()

	//1.create user
	user = &dao.User{
		Name:     LdapUserName,
		Password: LdapPassword,
		Email:    "lihanghua@huawei.com",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	SignUp(t, user)

	//2. create organization
	org = &dao.Organization{
		Name:            "huawei",
		Comment:         "create org",
		MemberPrivilege: dao.WRITE,
	}
	statusCode, err := CreateOrganization(t, org, user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create Organization Failed.")
	}
}

func Test_UserNamespaceMode(t *testing.T) {

	path := DockyardPath + "/conf/runtime.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}

	//user namespace mode
	conf.Set("auth_server::namespace", "user")
	conf.SaveConfigFile(path)
	if err := DockyardRestart(t); err != nil {
		t.Error(err)
	}

	if err := PushPullImage(user.Name, user); err != nil {
		t.Error(err)
	}
	if err := PushPullImage(org.Name, user); err != nil {
		t.Logf(" %v", err.Error())
	} else {
		t.Errorf("Test user namespace mode Failed")
	}
}

func Test_OrganizationNamespaceMode(t *testing.T) {

	path := DockyardPath + "/conf/runtime.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}

	//organization namespace mode
	conf.Set("auth_server::namespace", "organization")
	conf.SaveConfigFile(path)
	if err := DockyardRestart(t); err != nil {
		t.Error(err)
	}

	if err := PushPullImage(user.Name, user); err != nil {
		t.Logf(" %v", err.Error())
	} else {
		t.Errorf("Test organization namespace mode Failed")
	}
	if err := PushPullImage(org.Name, user); err != nil {
		t.Error(err)
	}
}

func Test_AllNamespaceMode(t *testing.T) {

	path := DockyardPath + "/conf/runtime.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}

	//all namespace mode
	conf.Set("auth_server::namespace", "all")
	conf.SaveConfigFile(path)
	if err := DockyardRestart(t); err != nil {
		t.Error(err)
	}
	if err := PushPullImage(user.Name, user); err != nil {
		t.Error(err)
	}
	if err := PushPullImage(org.Name, user); err != nil {
		t.Error(err)
	}
}

func Test_NamespaceSetClear(t *testing.T) {
	DeleteOrganization(t, org.Name, user.Name, user.Password)
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	DeleteUser(t, user.Name, sysAdmin.Name, sysAdmin.Password)
}
