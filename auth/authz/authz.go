package authz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/huawei-openlab/newdb/orm"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

//Authz Singleton, new by NewAuthenticator()
var Authz *Authorizer

//AuthorizerConfig struct
type AuthorizerConfig struct {
	Issuer         string
	PrivateKeyFile string
	Expiration     int64
	NameSpaceMode  string
	AuthnMode      string
}

//Authorizer plugin interface.
type Authorizer struct {
	config     *AuthorizerConfig
	privateKey libtrust.PrivateKey
	sigAlg     string
}

//NameSpaceMode
const (
	NameSpaceUser         = "user"
	NameSpaceOrganization = "organization"
	NameSpaceAll          = "all"
)

//NewAuthorizer new Authorizer object
func NewAuthorizer() (*Authorizer, error) {

	if setting.NameSpace != NameSpaceUser && setting.NameSpace != NameSpaceOrganization && setting.NameSpace != NameSpaceAll {
		return nil, fmt.Errorf("unsupport namespace mode")
	}

	config := &AuthorizerConfig{
		Issuer:         setting.Issuer,
		PrivateKeyFile: setting.PrivateKey,
		Expiration:     setting.Expiration,
		NameSpaceMode:  setting.NameSpace,
		AuthnMode:      setting.Authn,
	}

	var (
		err    error
		pk     libtrust.PrivateKey
		sigAlg string
	)
	pk, err = libtrust.LoadKeyFile(setting.PrivateKey)
	if err != nil {
		return nil, err
	} else {
		// Sign something dummy to find out which algorithm is used.
		_, sigAlg, err = pk.Sign(strings.NewReader("dummy"), 0)
		if err != nil {
			return nil, err
		}
	}

	return &Authorizer{config: config, privateKey: pk, sigAlg: sigAlg}, nil
}

//GetAuthorize Get Authorize
func (authz *Authorizer) GetAuthorize(userName, password, service, scope string) (int, []byte) {
	return authz.authorize(userName, password, service, scope, "GET", "")
}

//DeleteAuthorize Delete repo Authorize
func (authz *Authorizer) DeleteAuthorize(userName, password, service, scope, isdel string) (int, []byte) {
	return authz.authorize(userName, password, service, scope, "DELETE", isdel)
}

//PostAuthorize Post Authorize
func (authz *Authorizer) PostAuthorize(userName, password, service, scope string) (int, []byte) {
	return authz.authorize(userName, password, service, scope, "POST", "")
}

func errorToJSON(err error) []byte {
	errorCode := errcode.Errors{
		errcode.Error{
			Code:    errcode.ErrorCodeUnauthorized,
			Message: err.Error(),
		},
	}
	buf, _ := errorCode.MarshalJSON()
	return buf
}

func (authz *Authorizer) authorize(userName, password, service, scope, authzType, isdel string) (int, []byte) {

	//authn
	user, err := authn.Login(userName, password)
	if err != nil {
		return http.StatusUnauthorized, errorToJSON(err)
	}

	//get actions
	access, err := authz.getResourceActions(scope)
	if err != nil {
		return http.StatusBadRequest, errorToJSON(err)
	}
	for _, a := range access {
		if a.Type != "repository" {
			continue
		}
		if nameSpace, repo, err := authz.getNameSpaceRepositoryName(a.Name); err != nil {
			return http.StatusBadRequest, errorToJSON(err)
		} else {
			//get accesss
			var (
				err error
				as  []string
			)
			switch authzType {
			case "GET", "POST":
				as, err = authz.getAccess(user, nameSpace, repo, a.Actions)
			case "DELETE":
				as, err = authz.deleteAccess(user, nameSpace, repo, a.Actions, isdel)
			default:
				err = fmt.Errorf("not support authz type: %s", authzType)
			}
			if err != nil {
				if strings.HasPrefix(err.Error(), dao.StatusUnauthorized) {
					return http.StatusUnauthorized, errorToJSON(err)
				} else if strings.HasPrefix(err.Error(), dao.StatusBadRequest) {
					return http.StatusBadRequest, errorToJSON(err)
				} else if strings.HasPrefix(err.Error(), dao.StatusNotFound) {
					return http.StatusNotFound, errorToJSON(err)
				} else {
					return http.StatusInternalServerError, errorToJSON(err)
				}
			}
			a.Actions = as
		}
	}

	//create token
	t := &Token{
		UserName: userName,
		Service:  service,
		Access:   access,

		Issuer:     authz.config.Issuer,
		Expiration: authz.config.Expiration,

		PrivateKey: authz.privateKey,
		SigAlg:     authz.sigAlg,
	}
	rawToken, err := t.MakeToken()
	if err != nil {
		return http.StatusInternalServerError, errorToJSON(err)
	}

	//marshal
	tk := make(map[string]string)
	tk["token"] = rawToken
	if result, err := json.Marshal(tk); err != nil {
		return http.StatusInternalServerError, errorToJSON(err)
	} else {
		return http.StatusOK, result
	}
}

// GetResourceActions
func (authz *Authorizer) getResourceActions(scope string) ([]*token.ResourceActions, error) {
	res := []*token.ResourceActions{}
	if scope == "" {
		return res, nil
	}

	for _, resourceScope := range strings.Split(scope, " ") {
		items := strings.SplitN(resourceScope, ":", 3)
		if len(items) != 3 {
			return nil, fmt.Errorf("resourcescope format is not correct")
		} else if items[0] == "" || items[1] == "" || items[2] == "" {
			return nil, fmt.Errorf("resourcescope format is not correct")
		}
		res = append(res, &token.ResourceActions{
			Type:    items[0],
			Name:    items[1],
			Actions: strings.Split(items[2], ","),
		})
	}
	return res, nil
}

func (authz *Authorizer) getNameSpaceRepositoryName(name string) (string, string, error) {
	if strings.Contains(name, "/") {
		s := strings.LastIndex(name, "/")
		nameSpace := name[0:s]
		repo := name[s+1:]
		return nameSpace, repo, nil
	} else {
		return "", "", fmt.Errorf("scope format error")
	}
}

//return W,R
func (authz *Authorizer) getPermission(user *dao.User, nameSpace string, repo string, actions []string) (string, bool, error) {
	var (
		permission string
		//IsOrgRep org or user repo
		IsOrgRep = false
	)

	//1.check is organization or user namespace
	org := &dao.Organization{Name: nameSpace}
	if exist, err := org.Get(); err != nil {
		return "", false, err
	} else if exist {
		if org.Status == dao.ACTIVE {
			IsOrgRep = true
		} else {
			return "", false, fmt.Errorf(dao.StatusBadRequest + ": namespace is inactive")
		}
	} else {
		u := &dao.User{Name: nameSpace}
		if exist, err := u.Get(); err != nil {
			return "", false, err
		} else if exist {
			if u.Status == dao.ACTIVE {
				IsOrgRep = false
			} else {
				return "", false, fmt.Errorf(dao.StatusBadRequest + ": namespace is inactive")
			}
		} else {
			return "", false, fmt.Errorf(dao.StatusNotFound + ": namespace is not exist")
		}
	}

	//2.check namespace mode
	if authz.config.NameSpaceMode == NameSpaceUser && IsOrgRep {
		return "", false, fmt.Errorf(dao.StatusBadRequest + ": can get organizaiton namespace permission as config is user namespace mode")
	} else if authz.config.NameSpaceMode == NameSpaceOrganization && !IsOrgRep {
		return "", false, fmt.Errorf(dao.StatusBadRequest + ": can get user namespace permission as config is organizaiton namespace mode")
	}

	//3.get namespace's permission
	var err error
	if IsOrgRep {
		orgNSP := OrganizationNSPermission{
			User:               user,
			NameSpace:          nameSpace,
			Repo:               repo,
			OrgMemberPrivilege: org.MemberPrivilege,
			Actions:            actions,
		}
		permission, err = orgNSP.GetPermission()
	} else {
		userNSP := &UserNSPermission{
			User:      user,
			NameSpace: nameSpace,
			Repo:      repo,
			Actions:   actions,
		}
		permission, err = userNSP.GetPermission()
	}
	if err != nil {
		return "", IsOrgRep, err
	}
	return permission, IsOrgRep, nil
}

// return Actions = "push,pull", "pull", or ""
func (authz *Authorizer) getAccess(user *dao.User, nameSpace, repo string, actions []string) ([]string, error) {
	as := []string{}
	permission, _, err := authz.getPermission(user, nameSpace, repo, actions)
	if err != nil {
		return nil, err
	}

	if strings.Contains(permission, "ADMIN") || strings.Contains(permission, "W") {
		as = append(as, "push")
		as = append(as, "pull")
	} else if strings.Contains(permission, "R") {
		as = append(as, "pull")
	}
	return as, nil
}

// return Actions = "*" or ""
func (authz *Authorizer) deleteAccess(user *dao.User, nameSpace, repo string, actions []string, isdel string) ([]string, error) {
	as := []string{}
	if actions == nil || actions[0] != "*" {
		return nil, fmt.Errorf("scope format error")
	}

	permission, isOrgRep, err := authz.getPermission(user, nameSpace, repo, actions)
	if err != nil {
		return nil, err
	}

	if strings.Contains(permission, "ADMIN") {
		as = append(as, "*")
		var r *dao.RepositoryEx
		if isOrgRep {
			r = &dao.RepositoryEx{
				Name:     repo,
				IsOrgRep: isOrgRep,
				Org:      &dao.Organization{Name: nameSpace},
			}
		} else {
			r = &dao.RepositoryEx{
				Name:     repo,
				IsOrgRep: isOrgRep,
				User:     &dao.User{Name: nameSpace},
			}
		}

		if exist, err := r.Get(); err != nil {
			return nil, err
		} else if exist && (isdel == "true") {
			o := orm.NewOrm()
			if _, err := o.Delete(r); err != nil {
				return nil, err
			}
		}
	}
	return as, nil
}
