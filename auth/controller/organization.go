package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
)

//2.organization manager
//Create
//Delete
//Update
//Retreve

//POST
func CreateOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Create Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	if _, err := authn.Login(userName, password); err != nil {
		log.Error("[Create Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgnization info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Create Organization] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	org := &dao.Organization{}
	if err := json.Unmarshal(body, org); err != nil {
		log.Error("[Create Organization] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//3. create organizaion and set user as organization's admin. It's a transaction ops
	oum := &dao.OrganizationUserMap{
		User: &dao.User{Name: userName},
		Role: dao.ORGADMIN,
		Org:  &dao.Organization{Name: org.Name},
	}
	if err := dao.CreateOrganization(org, oum); err != nil {
		log.Error("[Create Organization] Failed to Create Organization to db: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//DELETE :organization
func DeleteOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Delete Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Delete Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get delete orgnization
	orgName := ctx.Params(":organization")

	//3. get permisson for removing organization
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Delete Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Delete Organization] Unauthorized to delete organization")
		return http.StatusUnauthorized, []byte("Unauthorized to delete organization")
	}

	//4. remove organizaiton
	org := &dao.Organization{Name: orgName}
	if err := org.Delete(); err != nil {
		log.Error("[Delete Organization] Failed to Delete Organization: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//Deactive :organization
func DeactiveOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get deactive orgnization
	orgName := ctx.Params(":organization")

	//3. get permisson for removing organization
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Deactive Organization] Failed to get user permission: %v", err.Error())
		return ConvertError(err)
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Deactive Organization] Unauthorized to deactive organization")
		return http.StatusUnauthorized, []byte("Unauthorized to deactive organization")
	}

	//4. deactive organizaiton
	org := &dao.Organization{Name: orgName}
	if err := org.Deactive(); err != nil {
		log.Error("[Deactive Organization] Failed to deactive Organization: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

// ActiveOrganization returns response of active according to the request
func ActiveOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active Organization] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get active orgnization
	orgName := ctx.Params(":organization")

	//3. get permisson for removing organization
	permisson, err := getOrganizaionPermission(user, orgName)
	if err != nil {
		log.Error("[Active Organization] Failed to get user permission: " + err.Error())
		return ConvertError(err)
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Active Organization] Unauthorized to active organization")
		return http.StatusUnauthorized, []byte("Unauthorized to active organization")
	}

	//4. active organizaiton
	org := &dao.Organization{Name: orgName}
	if err := org.Active(); err != nil {
		log.Error("[Active Organization] Failed to active Organization: " + err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//PUT
func UpdateOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgnization info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Organization] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//get update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Organization] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "name" && k != "email" && k != "comment" && k != "url" && k != "memberprivilege" && k != "location" {
			log.Error("[Update Organization] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	fields := []string{}
	org := &dao.Organization{}
	if val, ok := dat["name"]; ok {
		org.Name = val.(string)
	}
	if val, ok := dat["email"]; ok {
		org.Email = val.(string)
		fields = append(fields, "email")
	}
	if val, ok := dat["comment"]; ok {
		org.Comment = val.(string)
		fields = append(fields, "comment")
	}
	if val, ok := dat["url"]; ok {
		org.URL = val.(string)
		fields = append(fields, "url")
	}
	if val, ok := dat["memberprivilege"]; ok {
		org.MemberPrivilege = int(val.(float64))
		fields = append(fields, "memberprivilege")
	}
	if val, ok := dat["location"]; ok {
		org.Location = val.(string)
		fields = append(fields, "location")
	}

	//3. get permisson
	permisson, err := getOrganizaionPermission(user, org.Name)
	if err != nil {
		log.Error("[Update Organization] Failed to get user permission: %v", err.Error())
		return ConvertError(err)
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Update Organization] Unauthorized to Update Organization Info")
		return http.StatusUnauthorized, []byte(" Unauthorized to Update Organization Info")
	}
	if err := org.Update(fields...); err != nil {
		log.Error("[Update Organization] Failed to Update Organization to db: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//GET: organization list
func GetOrganizationList(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get Organization List] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	if _, err := authn.Login(userName, password); err != nil {
		log.Error("[Get Organization List] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get organization list
	org := &dao.Organization{}
	if orglist, err := org.List(); err != nil {
		log.Error("[Get Organization List] Failed to get organization list: %v", err.Error())
		return ConvertError(err)
	} else {
		if result, err := json.Marshal(orglist); err != nil {
			log.Error("[Get Organization List] Failed to marshal organization list: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}

func getOrganizaionPermission(user *dao.User, orgName string) (int, error) {
	//user is system admin
	if user.Role == dao.SYSADMIN {
		return user.Role, nil
	}

	//user is org admin
	oum := &dao.OrganizationUserMap{
		Org:  &dao.Organization{Name: orgName},
		User: &dao.User{Name: user.Name},
	}
	if exist, err := oum.Get(); err != nil {
		return 0, err
	} else if !exist {
		return 0, nil
	} else {
		return oum.Role, nil
	}
}

func ConvertError(err error) (errCode int, desc []byte) {
	if strings.HasPrefix(err.Error(), dao.StatusUnauthorized) {
		return http.StatusUnauthorized, []byte(err.Error())
	} else if strings.HasPrefix(err.Error(), dao.StatusBadRequest) {
		return http.StatusBadRequest, []byte(err.Error())
	} else if strings.HasPrefix(err.Error(), dao.StatusNotFound) {
		return http.StatusNotFound, []byte(err.Error())
	} else {
		return http.StatusInternalServerError, []byte(err.Error())
	}
}
