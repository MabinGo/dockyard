package ldaptest

import (
	"testing"

	"github.com/astaxie/beego/config"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
)

var (
	ldapUser, users *dao.User
	orgL, org       *dao.Organization
	conf            config.Configer
	path            string
)

func Test_LdapInit(t *testing.T) {
	authtest.GetConfig()

	//1.create user
	ldapUser = &dao.User{
		Name:     authtest.LdapUserName,
		Password: authtest.LdapPassword,
		Email:    "lihanghua@huawei.com",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, ldapUser)

	users = &dao.User{
		Name:     "other_user",
		Password: "abc123456",
		Email:    "users@huawei.com",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, users)

	orgL = &dao.Organization{
		Name:            "huawei",
		Comment:         "create org",
		MemberPrivilege: dao.WRITE,
	}

	//read config file
	var err error
	path = authtest.DockyardPath + "/conf/runtime.conf"
	conf, err = config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}
}

func ldapPushPullImage() error {

	//user push pull images
	if err := authtest.PushPullImage(ldapUser.Name, ldapUser); err != nil {
		return err
	}
	//organization push pull images
	if err := authtest.PushPullImage(orgL.Name, ldapUser); err != nil {
		return err
	}
	return nil
}

func Test_LdapPlain(t *testing.T) {

	//config transportmethod = plain,  addr = LGGAD40-DC.china.huawei.com:389
	conf.Set("auth_server::authn", "authn_ldap")
	conf.Set("authn_ldap::transportmethod", "plain")
	conf.Set("authn_ldap::addr", "LGGAD40-DC.china.huawei.com:389")
	conf.Set("authn_ldap::insecuretlsskipverify", "true")
	conf.SaveConfigFile(path)
	if err := authtest.DockyardRestart(t); err != nil {
		t.Error(err)
	}

	//Create Organization
	statusCode, err := authtest.CreateOrganization(t, orgL, ldapUser.Name, ldapUser.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create Organization Failed.")
	}

	if err := ldapPushPullImage(); err != nil {
		t.Error(err)
	}
}

func Test_LdapStarttls(t *testing.T) {

	//config transportmethod = starttls, addr = LGGAD40-DC.china.huawei.com:389
	conf.Set("auth_server::authn", "authn_ldap")
	conf.Set("authn_ldap::transportmethod", "starttls")
	conf.Set("authn_ldap::addr", "LGGAD40-DC.china.huawei.com:389")
	conf.Set("authn_ldap::insecuretlsskipverify", "true")
	conf.SaveConfigFile(path)
	if err := authtest.DockyardRestart(t); err != nil {
		t.Error(err)
	}

	if err := ldapPushPullImage(); err != nil {
		t.Error(err)
	}
}

func Test_LdapTls(t *testing.T) {

	//config transportmethod = tls,  addr = LGGAD40-DC.china.huawei.com:636
	conf.Set("auth_server::authn", "authn_ldap")
	conf.Set("authn_ldap::transportmethod", "tls")
	conf.Set("authn_ldap::addr", "LGGAD40-DC.china.huawei.com:636")
	conf.Set("authn_ldap::insecuretlsskipverify", "true")
	conf.SaveConfigFile(path)
	if err := authtest.DockyardRestart(t); err != nil {
		t.Error(err)
	}

	if err := ldapPushPullImage(); err != nil {
		t.Error(err)
	}
}

func Test_NonLdapUserCreateOrganization(t *testing.T) {

	//. create org organization
	org = &dao.Organization{
		Name:            "huaweicom",
		Comment:         "create org",
		MemberPrivilege: dao.WRITE,
	}
	statusCode, err := authtest.CreateOrganization(t, org, users.Name, users.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create Organization Error")
	}
}

func Test_NonLdapUserPushPullImage(t *testing.T) {

	//user push pull images
	if err := authtest.PushPullImage(users.Name, users); err != nil {
		t.Logf(" %v", err.Error())
	} else {
		t.Fatal("ldap authn Failed")
	}
}

func Test_LdapClear(t *testing.T) {

	//config authn = authn_db
	conf.Set("auth_server::authn", "authn_db")
	conf.SaveConfigFile(path)
	if err := authtest.DockyardRestart(t); err != nil {
		t.Error(err)
	}

	authtest.DeleteOrganization(t, orgL.Name, ldapUser.Name, ldapUser.Password)
	authtest.DeleteOrganization(t, org.Name, users.Name, users.Password)
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	authtest.DeleteUser(t, ldapUser.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, users.Name, sysAdmin.Name, sysAdmin.Password)
}
