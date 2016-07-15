package authn

import (
	"fmt"

	"github.com/containerops/dockyard/auth/authn/db"
	"github.com/containerops/dockyard/auth/authn/huaweiw3"
	"github.com/containerops/dockyard/auth/authn/ldap"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

//Authn Singleton, new by NewAuthenticator()
var Authn Authenticator

// Authentication plugin interface.
type Authenticator interface {
	Authenticate(userName, password string) (*dao.User, error)
}

const (
	Authn_db       = "authn_db"
	Authn_ldap     = "authn_ldap"
	Authn_huaweiw3 = "authn_huaweiw3"
)

func NewAuthenticator() (Authenticator, error) {
	switch setting.Authn {
	case Authn_db:
		return db.NewDBAuthn()
	case Authn_ldap:
		lConfig := &ldap.LDAPAuthnConfig{
			Addr:                  setting.Addr,
			TransportMethod:       setting.TransportMethod,
			BaseDN:                setting.BaseDN,
			Filter:                setting.Filter,
			BindDN:                setting.BindDN,
			BindPassword:          setting.BindPassword,
			InsecureTLSSkipVerify: setting.InsecureTLSSkipVerify,
			CertFile:              setting.CertFile,
		}
		return ldap.NewLDAPAuthn(lConfig)
	case Authn_huaweiw3:
		hConfig := &huaweiw3.HuaweiW3AuthnConfig{
			Addr: setting.Addr,
			InsecureTLSSkipVerify: setting.InsecureTLSSkipVerify,
			CertFile:              setting.CertFile,
		}
		return huaweiw3.NewHuaweiW3Authn(hConfig)
	default:
		return nil, fmt.Errorf("unrecognized authn mode: %s", setting.Authn)
	}
}

// Login authenticates user credentials based on setting.
func Login(userName, password string) (*dao.User, error) {
	if Authn == nil {
		return nil, fmt.Errorf("singleton authn.Authn should be instance")
	}

	// can use root login for all authn mode
	if userName == "root" {
		dbAuthn, _ := db.NewDBAuthn()
		return dbAuthn.Authenticate(userName, password)
	}

	return Authn.Authenticate(userName, password)

}
