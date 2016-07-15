package authtest

import (
	"testing"

	"github.com/astaxie/beego/config"

	"github.com/containerops/dockyard/auth/dao"
)

var (
	w3User, users *dao.User
	orgHw, orgSw  *dao.Organization
	conf          config.Configer
	path          string
)

func Test_HuaweiW3Init(t *testing.T) {
	GetConfig()

	//1.create user
	w3User = &dao.User{
		Name:     LdapUserName,
		Password: LdapPassword,
		Email:    "lihanghua@huawei.com",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	SignUp(t, w3User)

	users = &dao.User{
		Name:     "other_user",
		Password: "abc123456",
		Email:    "users@huawei.com",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	SignUp(t, users)

	orgHw = &dao.Organization{
		Name:            "huawei",
		Comment:         "create org",
		MemberPrivilege: dao.WRITE,
	}

	//read config file
	var err error
	path = DockyardPath + "/conf/runtime.conf"
	conf, err = config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}
}

func Test_HuaweiW3(t *testing.T) {

	//config authn = authn_huaweiw3, insecuretlsskipverify = false
	conf.Set("auth_server::authn", "authn_huaweiw3")
	conf.Set("authn_huaweiw3::insecuretlsskipverify", "false")
	conf.SaveConfigFile(path)
	if err := DockyardRestart(t); err != nil {
		t.Error(err)
	}

	//1. create organization
	statusCode, err := CreateOrganization(t, orgHw, w3User.Name, w3User.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Create Organization Failed.")
	}

	//2. user push pull images
	if err := PushPullImage(w3User.Name, w3User); err != nil {
		t.Error(err)
	}

	//3. organization push pull images
	if err := PushPullImage(orgHw.Name, w3User); err != nil {
		t.Error(err)
	}
}

func Test_NonW3UserCreateOrganization(t *testing.T) {

	//. create org organization
	orgSw = &dao.Organization{
		Name:            "software",
		Comment:         "create org",
		MemberPrivilege: dao.WRITE,
	}
	statusCode, err := CreateOrganization(t, orgSw, users.Name, users.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create Organization Error")
	}
}

func Test_NonW3UserPushPullImage(t *testing.T) {

	//user push pull images
	if err := PushPullImage(users.Name, users); err != nil {
		t.Logf(" %v", err.Error())
	} else {
		t.Fatal("huaweiw3 authn Failed")
	}
}

func Test_HuaweiW3Clear(t *testing.T) {
	//config authn = authn_db
	conf.Set("auth_server::authn", "authn_db")
	conf.SaveConfigFile(path)
	if err := DockyardRestart(t); err != nil {
		t.Error(err)
	}

	DeleteOrganization(t, orgHw.Name, w3User.Name, w3User.Password)
	DeleteOrganization(t, orgSw.Name, w3User.Name, w3User.Password)
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	DeleteUser(t, w3User.Name, sysAdmin.Name, sysAdmin.Password)
	DeleteUser(t, users.Name, sysAdmin.Name, sysAdmin.Password)
}
