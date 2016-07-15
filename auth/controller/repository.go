package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
)

type RepositoryJSON struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"ispublic"` //true:public; false:private
	Comment  string `json:"comment,omitempty"`
	OrgName  string `json:"orgname,omitempty"`
	UserName string `json:"username,omitempty"`
}

//organization admin can create org repository
//user can create his reposity
//create orgnization repository body:{name:ubuntu,ispublic:true,comment: this is a repo,  orgname:huawei}
//create user repository body:{name:ubuntu,ispublic:true,comment: this is a repo, username:liugenping}
func CreateRepository(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Create Repository] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Create Repository] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Create Repository] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	repoJSON := &RepositoryJSON{}
	if err := json.Unmarshal(body, repoJSON); err != nil {
		log.Error("[Create Repository] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	if repoJSON.Name == "" {
		log.Error("[Create Repository] Repository name value is empty")
		return http.StatusBadRequest, []byte("Repository name value is empty")
	}
	if (repoJSON.OrgName != "" && repoJSON.UserName != "") ||
		(repoJSON.OrgName == "" && repoJSON.UserName == "") {
		log.Error("[Create Repository] Orgname and username are not null or null at same time")
		return http.StatusBadRequest, []byte("orgname and username are not null or null at same time")
	}

	var isOrgRep bool
	//3. get permisson
	if repoJSON.OrgName != "" { //org repository
		permisson, err := getOrganizaionPermission(user, repoJSON.OrgName)
		if err != nil {
			log.Error("[Create Repository] Failed to get user permission: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
			log.Error("[Create Repository] Unauthorized to create repository")
			return http.StatusUnauthorized, []byte("Unauthorized to create repository")
		}
		isOrgRep = true
	} else { //user repository
		//user is system admin
		if user.Role == dao.SYSADMIN || repoJSON.UserName == userName {
		} else {
			log.Error("[Create Repository] Unauthorized to create repository")
			return http.StatusUnauthorized, []byte("Unauthorized to create repository")
		}
		isOrgRep = false
	}

	//4. save
	repo := &dao.RepositoryEx{
		Name:     repoJSON.Name,
		IsPublic: repoJSON.IsPublic,
		Comment:  repoJSON.Comment,
		IsOrgRep: isOrgRep,
		Org:      &dao.Organization{Name: repoJSON.OrgName},
		User:     &dao.User{Name: repoJSON.UserName},
	}
	if err := repo.Save(); err != nil {
		log.Error("[Create Repository] Failed to save repository to db: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//DELETE  :namespace/:repository
func DeleteRepository(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Delete Repository] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Delete Repository] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	namespace := ctx.Params(":namespace")
	repo := ctx.Params(":repository")

	r := &dao.RepositoryEx{}

	//3.check is organization or user namespace
	IsOrgRep := false
	org := &dao.Organization{Name: namespace}
	if exist, err := org.Get(); err != nil {
		log.Error("[Delete Repository] Failed to get organization: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if exist {
		IsOrgRep = true
	} else {
		u := &dao.User{Name: namespace}
		if exist, err := u.Get(); err != nil {
			log.Error("[Delete Repository] Failed to get user: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if exist {
			IsOrgRep = false
		} else {
			log.Error("[Delete Repository] Namespace is not exist")
			return http.StatusNotFound, []byte("Namespace is not exist")
		}
	}

	//4.get permission and delete
	if !IsOrgRep {
		if user.Role == dao.SYSADMIN || namespace == userName { //user repository
			r.Name = repo
			r.IsOrgRep = false
			r.User = &dao.User{Name: namespace}
		} else {
			log.Error("[Delete Repository] Unauthorized to delete repository")
			return http.StatusUnauthorized, []byte("Unauthorized to delete repository")
		}
	} else { //org repository,only system admin and organization admin can delete repository
		permisson, err := getOrganizaionPermission(user, namespace)
		if err != nil {
			log.Error("[Delete Repository] Failed to get user permission: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
			log.Error("[Delete Repository] Unauthorized to delete repository")
			return http.StatusUnauthorized, []byte("Unauthorized to delete repository")
		}
		r.Name = repo
		r.IsOrgRep = true
		r.Org = &dao.Organization{Name: namespace}
	}

	if err := r.Delete(); err != nil {
		log.Error("[Delete Repository] Delete db reocorde error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

//Deactive  :namespace/:repository
func DeactiveRepository(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive Repository] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive Repository] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	namespace := ctx.Params(":namespace")
	repo := ctx.Params(":repository")

	r := &dao.RepositoryEx{}

	//3.check is organization or user namespace
	IsOrgRep := false
	org := &dao.Organization{Name: namespace}
	if exist, err := org.Get(); err != nil {
		log.Error("[Deactive Repository] Failed to get organization: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if exist {
		IsOrgRep = true
	} else {
		u := &dao.User{Name: namespace}
		if exist, err := u.Get(); err != nil {
			log.Error("[Deactive Repository] Failed to get user: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if exist {
			IsOrgRep = false
		} else {
			log.Error("[Deactive Repository] Namespace is not exist")
			return http.StatusNotFound, []byte("Namespace is not exist")
		}
	}

	//4.get permission and deactive
	if !IsOrgRep {
		if user.Role == dao.SYSADMIN || namespace == userName { //user repository
			r.Name = repo
			r.IsOrgRep = false
			r.User = &dao.User{Name: namespace}
		} else {
			log.Error("[Deactive Repository] Unauthorized to deactive repository")
			return http.StatusUnauthorized, []byte("Unauthorized to deactive repository")
		}
	} else { //org repository,only system admin and organization admin can deactive repository
		permisson, err := getOrganizaionPermission(user, namespace)
		if err != nil {
			log.Error("[Deactive Repository] Failed to get user permission: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
			log.Error("[Deactive Repository] Unauthorized to deactive repository")
			return http.StatusUnauthorized, []byte("Unauthorized to deactive repository")
		}
		r.Name = repo
		r.IsOrgRep = true
		r.Org = &dao.Organization{Name: namespace}
	}

	if err := r.Deactive(); err != nil {
		log.Error("[Deactive Repository] Deactive db reocorde error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

// ActiveRepository returns response of active according to the request
func ActiveRepository(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active Repository] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active Repository] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	namespace := ctx.Params(":namespace")
	repo := ctx.Params(":repository")

	r := &dao.RepositoryEx{}

	//3.check is organization or user namespace
	IsOrgRep := false
	org := &dao.Organization{Name: namespace}
	if exist, err := org.Get(); err != nil {
		log.Error("[Active Repository] Failed to get organization: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if exist {
		IsOrgRep = true
	} else {
		u := &dao.User{Name: namespace}
		if exist, err := u.Get(); err != nil {
			log.Error("[Active Repository] Failed to get user: " + err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if exist {
			IsOrgRep = false
		} else {
			log.Error("[Active Repository] Namespace is not exist")
			return http.StatusNotFound, []byte("Namespace is not exist")
		}
	}

	//4.get permission and active
	if !IsOrgRep {
		if user.Role == dao.SYSADMIN || namespace == userName { //user repository
			r.Name = repo
			r.IsOrgRep = false
			r.User = &dao.User{Name: namespace}
		} else {
			log.Error("[Active Repository] Unauthorized to active repository")
			return http.StatusUnauthorized, []byte("Unauthorized to active repository")
		}
	} else { //org repository,only system admin and organization admin can active repository
		permisson, err := getOrganizaionPermission(user, namespace)
		if err != nil {
			log.Error("[Active Repository] Failed to get user permission: " + err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
			log.Error("[Active Repository] Unauthorized to active repository")
			return http.StatusUnauthorized, []byte("Unauthorized to active repository")
		}
		r.Name = repo
		r.IsOrgRep = true
		r.Org = &dao.Organization{Name: namespace}
	}

	if err := r.Active(); err != nil {
		log.Error("[Active Repository] Active db reocorde error: " + err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

//organization admin can update org repository
//user can update his reposity
//update orgnization repository body:{name:ubuntu,ispublic:true,comment: this is a repo,  orgname:huawei}
//update user repository body:{name:ubuntu,ispublic:true,comment: this is a repo, username:liugenping}
func UpdateRepository(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Repository] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Repository] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Repository] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//get update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Repository] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "name" && k != "ispublic" && k != "comment" && k != "orgname" && k != "username" {
			log.Error("[Update Repository] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	fields := []string{}
	repo := &dao.RepositoryEx{
		Org:  &dao.Organization{Name: ""},
		User: &dao.User{Name: ""},
	}
	if val, ok := dat["name"]; ok {
		repo.Name = val.(string)
	}
	if val, ok := dat["ispublic"]; ok {
		repo.IsPublic = val.(bool)
		fields = append(fields, "ispublic")
	}
	if val, ok := dat["comment"]; ok {
		repo.Comment = val.(string)
		fields = append(fields, "comment")
	}
	if val, ok := dat["orgname"]; ok {
		repo.Org = &dao.Organization{Name: val.(string)}
	}
	if val, ok := dat["username"]; ok {
		repo.User = &dao.User{Name: val.(string)}
	}

	if repo.Name == "" {
		log.Error("[Update Repository] Repository Name value is empty")
		return http.StatusBadRequest, []byte("Repository Name value is empty")
	}
	if (repo.Org.Name != "" && repo.User.Name != "") ||
		(repo.Org.Name == "" && repo.User.Name == "") {
		log.Error("[Update Repository] Orgname and username are not null or null at same time")
		return http.StatusBadRequest, []byte("orgname and username are not null or null at same time")
	}

	var isOrgRep bool
	//3. get permisson
	if repo.Org.Name != "" { //org repository
		permisson, err := getOrganizaionPermission(user, repo.Org.Name)
		if err != nil {
			log.Error("[Update Repository] Failed to get user permission: %v", err.Error())
			return ConvertError(err)
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
			log.Error("[Update Repository] Unauthorized to update repository")
			return http.StatusUnauthorized, []byte("Unauthorized to update repository")
		}
		isOrgRep = true
	} else { //user repository
		//user is system admin
		if user.Role == dao.SYSADMIN || repo.User.Name == userName {
		} else {
			log.Error("[Update Repository] Unauthorized to update repository")
			return http.StatusUnauthorized, []byte("Unauthorized to update repository")
		}
		isOrgRep = false
	}

	//4. update
	repo.IsOrgRep = isOrgRep
	if err := repo.Update(fields...); err != nil {
		log.Error("[Update Repository] Failed to update repository to db: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//GET: repository list from namespace
func GetRepositoryList(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get Repository List] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get Repository List] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	var isOrgRep bool
	namespace := ctx.Params(":namespace")
	//2. judge organization or user
	User := &dao.User{Name: namespace}
	Org := &dao.Organization{Name: namespace}
	if exist, err := User.Get(); err != nil {
		log.Error("[Get Repository List] Failed to get user in db: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if exist {
		isOrgRep = false
	} else if exist, err = Org.Get(); err != nil {
		log.Error("[Get Repository List] Failed to get organization in db: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if exist {
		isOrgRep = true
	} else {
		log.Error("[Get Repository List] Namespace is not exist")
		return http.StatusBadRequest, []byte("Namespace is not exist")
	}

	//3. get permission
	if isOrgRep == true {
		if permisson, err := getOrganizaionPermission(user, namespace); err != nil {
			log.Error("[Get Repository List] Failed to get user permission: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.ORGMEMBER {
			log.Error("[Get Repository List] Unauthorized to get repository list from organization")
			return http.StatusUnauthorized, []byte("Unauthorized to get repository list from organization")
		}
	}

	//4. get repository list from namespace
	var repo *dao.RepositoryEx
	if isOrgRep == true {
		repo = &dao.RepositoryEx{IsOrgRep: isOrgRep, Org: &dao.Organization{Name: namespace}}
	} else {
		repo = &dao.RepositoryEx{IsOrgRep: isOrgRep, User: &dao.User{Name: namespace}}
	}
	if repolist, err := repo.List(); err != nil {
		log.Error("[Get Repository List] Failed to get repository list from namespace: %v", err.Error())
		return ConvertError(err)
	} else {
		if result, err := json.Marshal(repolist); err != nil {
			log.Error("[Get Repository List] Failed to marshal repository list in namespace: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}

func GetFuzzyRepositoryList(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get Repository List  From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	if _, err := authn.Login(userName, password); err != nil {
		log.Error("[Get Repository List From Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	fuzzyrepo := ctx.Params(":repository")
	//2. get fuzzy repository list
	repo := &dao.RepositoryEx{Name: fuzzyrepo}
	if repolist, err := repo.FuzzyList(); err != nil {
		log.Error("[Get Fuzzy Repository List] Failed to get fuzzy repository list: %v", err.Error())
		return ConvertError(err)
	} else {
		if result, err := json.Marshal(repolist); err != nil {
			log.Error("[Get Repository List From Organization] Failed to marshal repository list in organization: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}
