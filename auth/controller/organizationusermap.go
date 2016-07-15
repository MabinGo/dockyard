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

type OrganizationUserMapJSON struct {
	UserName string `json:"username"`
	Role     int    `json:"role"` //role of user in organization, owner or member
	OrgName  string `json:"orgname"`
	Status   int    `json:"status,omitempty"` //status: active(0) or inactive(1)
}

func AddUserToOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Add User To Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Add User To Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get user, role, org
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Add User To Organization] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	oumJSON := &OrganizationUserMapJSON{}
	if err := json.Unmarshal(body, oumJSON); err != nil {
		log.Error("[Add User To Organization] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, oumJSON.OrgName)
	if err != nil {
		log.Error("[Add User To Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Add User To Organization] Unauthorized to add user to organization")
		return http.StatusUnauthorized, []byte(" Unauthorized to add user to organization")
	}

	//4. Save
	oum := &dao.OrganizationUserMap{
		User: &dao.User{Name: oumJSON.UserName},
		Role: oumJSON.Role,
		Org:  &dao.Organization{Name: oumJSON.OrgName},
	}
	if err := oum.Save(); err != nil {
		log.Error("[Add User To Organization] Save db reocorde error:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, nil
}

func RemoveUserFromOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Remove User From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Remove User From Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get org name and user name
	orgName := ctx.Params(":organization")
	uName := ctx.Params(":user")

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Remove User From Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Remove User From Organization] Unauthorized to remove user from organization")
		return http.StatusUnauthorized, []byte("Unauthorized to remove user from organization")
	}

	//4. delete
	orgUserMap := &dao.OrganizationUserMap{
		Org:  &dao.Organization{Name: orgName},
		User: &dao.User{Name: uName},
	}
	if err := orgUserMap.Delete(); err != nil {
		log.Error("[Remove User From Organization] Failed to Remove User: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, nil
}

// DeactiveUserFromOrganization returns response of deactive according to the request
func DeactiveUserFromOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive User From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive User From Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get org name and user name
	orgName := ctx.Params(":organization")
	uName := ctx.Params(":user")

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Deactive User From Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Deactive User From Organization] Unauthorized to deactive user from organization")
		return http.StatusUnauthorized, []byte("Unauthorized to deactive user from organization")
	}

	//4. deactive
	orgUserMap := &dao.OrganizationUserMap{
		Org:  &dao.Organization{Name: orgName},
		User: &dao.User{Name: uName},
	}
	if err := orgUserMap.Deactive(); err != nil {
		log.Error("[Deactive User From Organization] Failed to deactive User: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, nil
}

// ActiveUserFromOrganization returns response of active according to the request
func ActiveUserFromOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active User From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active User From Organization] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get org name and user name
	orgName := ctx.Params(":organization")
	uName := ctx.Params(":user")

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Active User From Organization] Failed to get user permission: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Active User From Organization] Unauthorized to active user from organization")
		return http.StatusUnauthorized, []byte("Unauthorized to active user from organization")
	}

	//4. active
	orgUserMap := &dao.OrganizationUserMap{
		Org:  &dao.Organization{Name: orgName},
		User: &dao.User{Name: uName},
	}
	if err := orgUserMap.Active(); err != nil {
		log.Error("[Active User From Organization] Failed to active User: " + err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, nil
}

func UpdateOrganizationUserMap(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Organization User Map] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Organization User Map] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get user, role, org
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Organization User Map] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	oumJSON := &OrganizationUserMapJSON{}
	if err := json.Unmarshal(body, oumJSON); err != nil {
		log.Error("[Update Organization User Map] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//check update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Organization User Map] Failed to unmarshal request body")
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "username" && k != "role" && k != "orgname" {
			log.Error("[Update Organization User Map] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, oumJSON.OrgName)
	if err != nil {
		log.Error("[Update Organization User Map] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Update Organization User Map] Unauthorized to update organization user map info")
		return http.StatusUnauthorized, []byte(" Unauthorized to update organization user map info")
	}

	//4. Update
	oum := &dao.OrganizationUserMap{
		User: &dao.User{Name: oumJSON.UserName},
		Role: oumJSON.Role,
		Org:  &dao.Organization{Name: oumJSON.OrgName},
	}
	if err := oum.Update("Role"); err != nil {
		log.Error("[Update Organization User Map] Update db reocorde error:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, nil
}

func GetUserListFromOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get User List From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get User List From Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get permisson
	orgName := ctx.Params(":organization")
	if permisson, err := getOrganizaionPermission(user, orgName); err != nil {
		log.Error("[Get User List From Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Get User List From Organization] Unauthorized to get user list from organization")
		return http.StatusUnauthorized, []byte("Unauthorized to get user list from organization")
	}

	//3. get user list from organization
	oum := &dao.OrganizationUserMap{Org: &dao.Organization{Name: orgName}}
	if oumlist, err := oum.List(); err != nil {
		log.Error("[Get User List From Organization] Failed to get user list from organization: %v", err.Error())
		return ConvertError(err)
	} else {
		oumjsonlist := []OrganizationUserMapJSON{}
		for _, eachoum := range oumlist {
			eachoumjson := OrganizationUserMapJSON{
				UserName: eachoum.User.Name,
				Role:     eachoum.Role,
				OrgName:  orgName,
				Status:   eachoum.Status,
			}
			oumjsonlist = append(oumjsonlist, eachoumjson)
		}
		if result, err := json.Marshal(oumjsonlist); err != nil {
			log.Error("[Get User List From Organization] Failed to marshal user list in organization: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}
