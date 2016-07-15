package ldap

import (
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
)

var (
	correctPassword string
)

func init() {
	correctPassword = "!"
}

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

func Test_AuthenticatePlain(t *testing.T) {

	conf := &LDAPAuthnConfig{
		Addr:                  "LGGAD40-DC.china.huawei.com:389",
		TransportMethod:       "plain",
		BaseDN:                "dc=china,dc=huawei,dc=com",
		Filter:                "(&(sAMAccountName=${account})(objectClass=person))",
		BindDN:                "CN=liugenping 00257029,OU=CorpUsers,DC=china,DC=huawei,DC=com",
		BindPassword:          correctPassword,
		InsecureTLSSkipVerify: true,
		CertFile:              "",
	}

	la, err := NewLDAPAuthn(conf)
	if err != nil {
		t.Error(err)
	}

	if _, err := la.Authenticate("l00257029", correctPassword); err != nil {
		t.Error(err)
	}
}

func Test_AuthenticateTLSSkipVerify(t *testing.T) {

	conf := &LDAPAuthnConfig{
		Addr:                  "LGGAD40-DC.china.huawei.com:636",
		TransportMethod:       "tls",
		BaseDN:                "dc=china,dc=huawei,dc=com",
		Filter:                "(&(sAMAccountName=${account})(objectClass=person))",
		BindDN:                "CN=liugenping 00257029,OU=CorpUsers,DC=china,DC=huawei,DC=com",
		BindPassword:          correctPassword,
		InsecureTLSSkipVerify: true,
		CertFile:              "",
	}

	la, err := NewLDAPAuthn(conf)
	if err != nil {
		t.Error(err)
	}

	if _, err := la.Authenticate("l00257029", correctPassword); err != nil {
		t.Error(err)
	}
}

func Test_AuthenticateStartTLSSkipVerify(t *testing.T) {

	conf := &LDAPAuthnConfig{
		Addr:                  "LGGAD40-DC.china.huawei.com:389",
		TransportMethod:       "starttls",
		BaseDN:                "dc=china,dc=huawei,dc=com",
		Filter:                "(&(sAMAccountName=${account})(objectClass=person))",
		BindDN:                "CN=liugenping 00257029,OU=CorpUsers,DC=china,DC=huawei,DC=com",
		BindPassword:          correctPassword,
		InsecureTLSSkipVerify: true,
		CertFile:              "",
	}

	la, err := NewLDAPAuthn(conf)
	if err != nil {
		t.Error(err)
	}

	if _, err := la.Authenticate("l00257029", correctPassword); err != nil {
		t.Error(err)
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
