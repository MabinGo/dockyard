package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-ldap/ldap"

	"github.com/containerops/dockyard/auth/dao"
)

type LDAPAuthnConfig struct {
	Addr                  string
	TransportMethod       string
	BaseDN                string
	Filter                string
	BindDN                string
	BindPassword          string
	InsecureTLSSkipVerify bool
	CertFile              string
}

type LDAPAuthn struct {
	config   *LDAPAuthnConfig
	certPool *x509.CertPool
}

func NewLDAPAuthn(config *LDAPAuthnConfig) (*LDAPAuthn, error) {
	switch config.TransportMethod {
	case "tls":
	case "starttls":
	case "plain":
	default:
		return nil, fmt.Errorf("unrecognized transport method: %s", config.TransportMethod)
	}

	var pool *x509.CertPool
	//TLS with verify cert
	if config.TransportMethod != "plain" && !config.InsecureTLSSkipVerify {
		pool = x509.NewCertPool()

		caCrt, err := ioutil.ReadFile(config.CertFile)
		if err != nil {
			return nil, err
		}
		if ok := pool.AppendCertsFromPEM(caCrt); !ok {
			return nil, fmt.Errorf("parse certificate error")
		}
	}

	l := &LDAPAuthn{
		config:   config,
		certPool: pool,
	}

	return l, nil
}

func (l *LDAPAuthn) Authenticate(userName, password string) (*dao.User, error) {

	//1.check if user exist in db
	user := &dao.User{Name: userName}
	if exist, err := user.Get(); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("Not found user:%s", userName)
	} else if exist && user.Status == dao.INACTIVE {
		return nil, fmt.Errorf("User:%s is inactive", userName)
	}

	//2.ldap authenticate
	conn, err := l.ldapConnect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	//err = conn.Bind("CN=liugenping 00257029,OU=CorpUsers,DC=china,DC=huawei,DC=com", password)
	if err := conn.Bind(l.config.BindDN, l.config.BindPassword); err != nil {
		return nil, err
	}

	uName := escapeAccount(userName)

	dn, err := l.ldapSearch(conn, uName)
	if err != nil {
		return nil, err
	}

	if err := conn.Bind(dn, password); err != nil {
		return nil, err
	}

	return user, nil
}

func (l *LDAPAuthn) ldapConnect() (*ldap.Conn, error) {
	var (
		conn *ldap.Conn
		err  error
	)
	s := strings.LastIndex(l.config.Addr, ":")
	tlsConf := &tls.Config{
		InsecureSkipVerify: l.config.InsecureTLSSkipVerify,
		RootCAs:            l.certPool,
		ServerName:         l.config.Addr[0:s],
	}

	switch l.config.TransportMethod {
	case "tls":
		conn, err = ldap.DialTLS("tcp", l.config.Addr, tlsConf)
	case "starttls":
		conn, err = ldap.Dial("tcp", l.config.Addr)
		if err != nil {
			err = conn.StartTLS(tlsConf)
		}
	case "plain":
		conn, err = ldap.Dial("tcp", l.config.Addr)
	default:
		return nil, fmt.Errorf("unrecognized transport method: %s", l.config.TransportMethod)
	}
	if err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

//ldap search and return required DN
func (l *LDAPAuthn) ldapSearch(conn *ldap.Conn, account string) (string, error) {

	//set filter
	filter := strings.NewReplacer("${account}", account).Replace(l.config.Filter)

	searchRequest := ldap.NewSearchRequest(
		l.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		nil,
		nil)
	sr, err := conn.Search(searchRequest)
	if err != nil {
		return "", err
	}
	if len(sr.Entries) != 1 {
		return "", fmt.Errorf("user does not exist or too many entries returned.")
	}
	return sr.Entries[0].DN, nil
}

//To prevent LDAP injection, some characters must be escaped for searching
//e.g. char '\' will be replaced by hex '\5c'
//Filter meta chars are choosen based on filter complier code
func escapeAccount(account string) string {
	r := strings.NewReplacer(
		`\`, `\5c`,
		`(`, `\28`,
		`)`, `\29`,
		`!`, `\21`,
		`*`, `\2a`,
		`&`, `\26`,
		`|`, `\7c`,
		`=`, `\3d`,
		`>`, `\3e`,
		`<`, `\3c`,
		`~`, `\7e`,
	)
	return r.Replace(account)
}
