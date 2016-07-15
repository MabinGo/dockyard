package huaweiw3

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containerops/dockyard/auth/dao"
)

type HuaweiW3AuthnConfig struct {
	Addr                  string
	InsecureTLSSkipVerify bool
	CertFile              string
}

type HuaweiW3Authn struct {
	config   *HuaweiW3AuthnConfig
	certPool *x509.CertPool
}

type W3LDAP struct {
	UserName   string `json:"userName"`
	Password   string `json:"password"`
	AuthMethod string `json:"authMethod"`
	Redirect   string `json:"redirect"`
}

func NewHuaweiW3Authn(config *HuaweiW3AuthnConfig) (*HuaweiW3Authn, error) {

	var pool *x509.CertPool = nil
	//TLS with verify cert
	if !config.InsecureTLSSkipVerify {
		pool = x509.NewCertPool()

		caCrt, err := ioutil.ReadFile(config.CertFile)
		if err != nil {
			return nil, err
		}
		if ok := pool.AppendCertsFromPEM(caCrt); !ok {
			return nil, fmt.Errorf("parse certificate error")
		}
	}

	ha := &HuaweiW3Authn{
		config:   config,
		certPool: pool,
	}
	return ha, nil
}

func (h *HuaweiW3Authn) Authenticate(userName, password string) (*dao.User, error) {

	//1.check if user exist in db
	user := &dao.User{Name: userName}
	if exist, err := user.Get(); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("Not found user:%s", userName)
	} else if exist && user.Status == dao.INACTIVE {
		return nil, fmt.Errorf("User:%s is inactive", userName)
	}

	//2.PSOT to w3 ldap
	w3 := &W3LDAP{
		UserName:   userName,
		Password:   password,
		AuthMethod: "password",
		Redirect:   "http://w3.huawei.com",
	}

	body, err := json.Marshal(w3)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", h.config.Addr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	//TLS with verify cert or not
	tlsConf := &tls.Config{
		InsecureSkipVerify: h.config.InsecureTLSSkipVerify,
		RootCAs:            h.certPool,
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConf,
		//DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("huawei w3 ldap return %d", resp.StatusCode)
	}

	return user, nil
}
