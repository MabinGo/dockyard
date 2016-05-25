package authz

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/huawei-openlab/newdb/orm"

	"github.com/containerops/dockyard/auth/authn"
	_ "github.com/containerops/dockyard/auth/authn/db"
	_ "github.com/containerops/dockyard/auth/authn/ldap"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	pk     libtrust.PrivateKey
	sigAlg string
	isInit bool
)

func AuthorizerOpen() error {
	var err error
	pk, err = libtrust.LoadKeyFile(setting.PrivateKey)
	if err != nil {
		isInit = false
		return err
	} else {
		// Sign something dummy to find out which algorithm is used.
		_, sigAlg, err = pk.Sign(strings.NewReader("dummy"), 0)
		if err != nil {
			isInit = false
			return err
		} else {
			isInit = true
			return nil
		}
	}
}

func GetAuthorize(userName, password, service, scope string) (int, []byte) {
	return authorize(userName, password, service, scope, "GET", "")
}

func DeleteAuthorize(userName, password, service, scope, isdel string) (int, []byte) {
	return authorize(userName, password, service, scope, "DELETE", isdel)
}

func PostAuthorize(userName, password, service, scope string) (int, []byte) {
	return authorize(userName, password, service, scope, "POST", "")
}

func authorize(userName, password, service, scope, authzType, isdel string) (int, []byte) {

	if !isInit {
		return http.StatusInternalServerError, []byte("private key error")
	}

	user, err := authn.Login(userName, password)
	if err != nil {
		return http.StatusUnauthorized, []byte(err.Error())
	}

	access, err := getResourceActions(scope)
	if err != nil {
		return http.StatusBadRequest, []byte(err.Error())
	}
	for _, a := range access {
		if a.Type != "repository" {
			continue
		}
		if nameSpace, repo, err := getNameSpaceRepositoryName(a.Name); err != nil {
			return http.StatusBadRequest, []byte(err.Error())
		} else {
			//get accesss
			var (
				err error
				as  []string
			)

			switch authzType {
			case "GET", "POST":
				as, err = getAccess(user, nameSpace, repo, a.Actions)
			case "DELETE":
				as, err = deleteAccess(user, nameSpace, repo, a.Actions, isdel)
			default:
				err = fmt.Errorf("not support authz type")
			}

			if err != nil {
				return http.StatusBadRequest, []byte(err.Error())
			}
			a.Actions = as
		}
	}
	//create token
	rawToken, err := makeToken(userName, service, access)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	//marshal
	tk := make(map[string]string)
	tk["token"] = rawToken
	if result, err := json.Marshal(tk); err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	} else {
		return http.StatusOK, result
	}
}

// GetResourceActions
func getResourceActions(scope string) ([]*token.ResourceActions, error) {
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

func getNameSpaceRepositoryName(name string) (string, string, error) {
	if strings.Contains(name, "/") {
		s := strings.LastIndex(name, "/")
		nameSpace := name[0:s]
		repo := name[s+1:]
		return nameSpace, repo, nil
	} else {
		return "", "", fmt.Errorf("scope format error")
	}
}

//if namespace is user name, namespace permission is W,else is defined by repository
func getUserNameSpacePermission(user *dao.User, nameSpace string, repo string, actions []string) (string, error) {
	permission := ""
	canCreateRepo := false
	IsRepoExist := false

	//only push can create repo if repo is not exit
	for _, a := range actions {
		if a == "push" {
			canCreateRepo = true
		}
	}

	if user.Name == nameSpace || user.Role == dao.SYSADMIN {
		permission = "ADMIN"
		canCreateRepo = canCreateRepo && true
	} else {
		permission = ""
		canCreateRepo = false
	}

	//delete repo
	if actions[0] == "*" {
		return permission, nil
	}

	//get repo privilege,create it if repo is not exist
	r := dao.RepositoryEx{
		Name:     repo,
		IsOrgRep: false,
		User:     &dao.User{Name: nameSpace},
	}
	if exist, err := r.Get(); err != nil {
		return "", err
	} else if exist {
		IsRepoExist = true
	} else {
		IsRepoExist = false
	}

	if IsRepoExist {
		if r.IsPublic {
			permission = permission + "R"
		}
	} else if !IsRepoExist && canCreateRepo {
		r.IsPublic = true
		if err := r.Save(); err != nil {
			return "", err
		}
		permission = permission + "R"
	} else {
		return "", fmt.Errorf("repository is not exist and login user didn't have privilege to create it")
	}

	return permission, nil
}

//repo is associated with team
func getTeamRepoPermission(userName string, repoID int) (string, error) {
	permission := ""

	if teamPermit, err := dao.GetTeamRepoPermit(repoID, userName); err != nil {
		return "", nil
	} else {
		for _, p := range teamPermit {
			if p == dao.WRITE {
				permission = permission + "W"
			} else {
				permission = permission + "R"
			}
		}
	}
	return permission, nil
}

//if namespace is org name,and login user is in organizaion:
//user is org's admin, then namespace permission is W.
//user is org's member, then namespace permission is same with  member's privilege
//else is defined by repository
func getOrgNameSpacePermission(user *dao.User, nameSpace string, repo string, memberPrivilege int, actions []string) (string, error) {
	permission := ""
	canCreateRepo := false
	IsRepoExist := false

	//only push can create repo if repo is not exit
	for _, a := range actions {
		if a == "push" {
			canCreateRepo = true
		}
	}

	if user.Role == dao.SYSADMIN {
		canCreateRepo = canCreateRepo && true
		permission = "ADMIN"
	} else {
		//get org user role
		orgUserMap := &dao.OrganizationUserMap{
			Org:  &dao.Organization{Name: nameSpace},
			User: &dao.User{Name: user.Name},
		}
		exist, err := orgUserMap.Get()
		if err != nil {
			return "", err
		}
		//user in organization
		if exist {
			//if org admin, permission is W
			if orgUserMap.Role == dao.ORGADMIN {
				canCreateRepo = canCreateRepo && true
				permission = "ADMIN"
			} else {
				canCreateRepo = false
				//memberPrivilege
				switch memberPrivilege {
				//case dao.ADMIN:
				//	permission = "W"
				case dao.WRITE:
					permission = "W"
				case dao.READ:
					permission = "R"
				case dao.NONE:
					permission = ""
				default:
					return "", fmt.Errorf("user's organization privilege is error")
				}
			}
		} else {
			permission = ""
			canCreateRepo = canCreateRepo && false
		}
	}

	//delete repo
	if actions[0] == "*" {
		return permission, nil
	}

	//get repo privilege,create it if repo is not exist
	r := dao.RepositoryEx{
		Name:     repo,
		IsOrgRep: true,
		Org:      &dao.Organization{Name: nameSpace},
	}
	if exist, err := r.Get(); err != nil {
		return "", err
	} else if exist {
		IsRepoExist = true
	} else {
		IsRepoExist = false
	}

	if IsRepoExist {
		if r.IsPublic {
			permission = permission + "R"
		}
		//add team perimit
		if p, err := getTeamRepoPermission(user.Name, r.Id); err != nil {
			return "", err
		} else {
			permission = permission + p
		}
	} else if !IsRepoExist && canCreateRepo {
		r.IsPublic = true
		if err := r.Save(); err != nil {
			return "", err
		}
		permission = permission + "R"
	} else {
		return "", fmt.Errorf("repository is not exist and login user didn't have privilege to create it")
	}

	return permission, nil
}

//return W,R
func getPermission(user *dao.User, nameSpace string, repo string, actions []string) (string, bool, error) {
	var (
		permission string
		IsOrgRep   bool = false
	)

	//1.check is organization or user namespace
	org := &dao.Organization{Name: nameSpace}
	if exist, err := org.Get(); err != nil {
		return "", false, err
	} else if exist {
		IsOrgRep = true
	} else {
		u := &dao.User{Name: nameSpace}
		if exist, err := u.Get(); err != nil {
			return "", false, err
		} else if exist {
			IsOrgRep = false
		} else {
			return "", false, fmt.Errorf("namespace is not exist")
		}
	}

	//2.get namespace's permission
	var err error
	if IsOrgRep {
		permission, err = getOrgNameSpacePermission(user, nameSpace, repo, org.MemberPrivilege, actions)
	} else {
		permission, err = getUserNameSpacePermission(user, nameSpace, repo, actions)
	}
	if err != nil {
		return "", IsOrgRep, err
	}
	return permission, IsOrgRep, nil
}

// return Actions = "push,pull", "pull", or ""
func getAccess(user *dao.User, nameSpace, repo string, actions []string) ([]string, error) {
	as := []string{}
	permission, _, err := getPermission(user, nameSpace, repo, actions)
	if err != nil {
		return nil, fmt.Errorf("Error occurred in GetPermission: %v", err)
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
func deleteAccess(user *dao.User, nameSpace, repo string, actions []string, isdel string) ([]string, error) {
	as := []string{}
	if actions == nil || actions[0] != "*" {
		return nil, fmt.Errorf("scope format error")
	}

	permission, isOrgRep, err := getPermission(user, nameSpace, repo, actions)
	if err != nil {
		return nil, fmt.Errorf("Error occurred in GetPermission: %v", err)
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

// MakeToken makes a valid jwt token based on parms.
func makeToken(username, service string, access []*token.ResourceActions) (string, error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: sigAlg,
		KeyID:      pk.KeyID(),
	}

	jwtID, err := randString(16)
	if err != nil {
		return "", fmt.Errorf("error to generate jwt id: %s", err)
	}

	now := time.Now().Unix()

	claimSet := &token.ClaimSet{
		Issuer:     setting.Issuer,
		Subject:    username,
		Audience:   service,
		Expiration: now + setting.Expiration,
		NotBefore:  now,
		IssuedAt:   now,
		JWTID:      jwtID,
		Access:     access,
	}

	var joseHeaderBytes, claimSetBytes []byte

	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		return "", fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return "", fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = pk.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return "", fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	return tokenString, nil
	//tk := token.NewToken(tokenString)
	//rs := fmt.Sprintf("%s.%s", tk.Raw, base64UrlEncode(tk.Signature))
	//return rs, nil
}

func randString(length int) (string, error) {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	for i, b := range rb {
		rb[i] = alphanum[int(b)%len(alphanum)]
	}
	return string(rb), nil
}

func base64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
