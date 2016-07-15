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

type TeamUserMapJSON struct {
	TeamName string `json:"teamname"`
	OrgName  string `json:"orgname"`
	UserName string `json:"username"`
	Role     int    `json:"role"`             //role of repository in team, read or write
	Status   int    `json:"status,omitempty"` //status: active(0) or inactive(1)
}

func AddUserToTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Add User To Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Add User To Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Add User To Team] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	tumJSON := &TeamUserMapJSON{}
	if err := json.Unmarshal(body, tumJSON); err != nil {
		log.Error("[Add User To Team] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	if tumJSON.TeamName == "" || tumJSON.UserName == "" || tumJSON.OrgName == "" {
		log.Error("[Add User To Team] Team name , organization name and user name can not be empty")
		return http.StatusUnauthorized, []byte("Team name , organization name and repository name can not be empty")
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, tumJSON.OrgName, tumJSON.TeamName)
	if err != nil {
		log.Error("[Add User To Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Add User To Team] Unauthorized to add user to team")
		return http.StatusUnauthorized, []byte("Unauthorized to add user to team")
	}

	//4. save
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: tumJSON.TeamName,
			Org:  &dao.Organization{Name: tumJSON.OrgName}},
		User: &dao.User{Name: tumJSON.UserName},
		Role: tumJSON.Role,
	}
	if err := tum.Save(); err != nil {
		log.Error("[Add User To Team] Failed to save to db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

func RemoveUserFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Remove User From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Remove User From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	uName := ctx.Params(":user")

	if teamName == "" || userName == "" || orgName == "" {
		log.Error("[Remove User From Team] Team name , organization name and user name can not be empty")
		return http.StatusUnauthorized, []byte("Team name , organization name and repository name can not be empty")
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Remove User From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Remove User From Team] Unauthorized to delete user from team")
		return http.StatusUnauthorized, []byte("Unauthorized to delete user from team")
	}

	//4. delete
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName}},
		User: &dao.User{Name: uName},
	}
	if err := tum.Delete(); err != nil {
		log.Error("[Delete User From Team] Failed to delete from db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

func DeactiveUserFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive User From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive User From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	uName := ctx.Params(":user")

	if teamName == "" || userName == "" || orgName == "" {
		log.Error("[Deactive User From Team] Team name , organization name and user name can not be empty")
		return http.StatusUnauthorized, []byte("Team name , organization name and repository name can not be empty")
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Deactive User From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Deactive User From Team] Unauthorized to deactive user from team")
		return http.StatusUnauthorized, []byte("Unauthorized to deactive user from team")
	}

	//4. deactive
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName}},
		User: &dao.User{Name: uName},
	}
	if err := tum.Deactive(); err != nil {
		log.Error("[Deactive User From Team] Failed to deactive from db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

// ActiveUserFromTeam returns response of active according to the request
func ActiveUserFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active User From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active User From Team] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	uName := ctx.Params(":user")

	if teamName == "" || userName == "" || orgName == "" {
		log.Error("[Active User From Team] Team name , organization name and user name can not be empty")
		return http.StatusUnauthorized, []byte("Team name , organization name and repository name can not be empty")
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Active User From Team] Failed to get user permission: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Active User From Team] Unauthorized to active user from team")
		return http.StatusUnauthorized, []byte("Unauthorized to active user from team")
	}

	//4. active
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName}},
		User: &dao.User{Name: uName},
	}
	if err := tum.Active(); err != nil {
		log.Error("[Active User From Team] Failed to active from db: " + err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//Get: user list from team
func GetUserListFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get User List From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get User List From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")

	//2. get permisson
	if permisson, err := getRepositoryTeamPermission(user, orgName, teamName); err != nil {
		log.Error("[Get User List From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Get User List From Team] Unauthorized to get user list from team")
		return http.StatusUnauthorized, []byte("Unauthorized to get user list from team")
	}

	//3. get user list from team
	tum := &dao.TeamUserMap{Team: &dao.Team{Name: teamName, Org: &dao.Organization{Name: orgName}}}
	if tumlist, err := tum.List(); err != nil {
		log.Error("[Get User List From Team] Failed to get user list from team: %v", err.Error())
		return ConvertError(err)
	} else {
		tumjsonlist := []TeamUserMapJSON{}
		for _, eachtum := range tumlist {
			eachtumjson := TeamUserMapJSON{
				TeamName: teamName,
				OrgName:  orgName,
				UserName: eachtum.User.Name,
				Role:     eachtum.Role,
				Status:   eachtum.Status,
			}
			tumjsonlist = append(tumjsonlist, eachtumjson)
		}
		if result, err := json.Marshal(tumjsonlist); err != nil {
			log.Error("[Get User List From Team] Failed to marshal user list in Team: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}

func UpdateTeamUserMap(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Team User Map] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Team User Map] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Team User Map] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	tumJSON := &TeamUserMapJSON{}
	if err := json.Unmarshal(body, tumJSON); err != nil {
		log.Error("[Update Team User Map] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	if tumJSON.TeamName == "" || tumJSON.UserName == "" || tumJSON.OrgName == "" {
		log.Error("[Update Team User Map] Team name , organization name and user name can not be empty")
		return http.StatusBadRequest, []byte("Team name , organization name and repository name can not be empty")
	}

	//check update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Team User Map] Failed to unmarshal request body")
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "teamname" && k != "orgname" && k != "username" && k != "role" {
			log.Error("[Update Team User Map] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, tumJSON.OrgName, tumJSON.TeamName)
	if err != nil {
		log.Error("[Update Team User Map] Failed to get user permission: %v", err.Error())
		return ConvertError(err)
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Update Team User Map] Unauthorized to update team user map info")
		return http.StatusUnauthorized, []byte("Unauthorized to update team user map info")
	}

	//4. update
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: tumJSON.TeamName,
			Org:  &dao.Organization{Name: tumJSON.OrgName}},
		User: &dao.User{Name: tumJSON.UserName},
		Role: tumJSON.Role,
	}
	if err := tum.Update("Role"); err != nil {
		log.Error("[Update Team User Map] Failed to update team user map info to db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}
