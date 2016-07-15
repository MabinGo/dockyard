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

type TeamJSON struct {
	TeamName string `json:"teamname"`
	Comment  string `json:"comment,omitempty"`
	OrgName  string `json:"orgname"`
	Status   int    `json:"status,omitempty"` //status: active(0) or inactive(1)
}

//team manager
//Create
//Delete
//Update
//Retreve

//POST
func CreateTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Create Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Create Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Create Team] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	teamJSON := &TeamJSON{}
	if err := json.Unmarshal(body, teamJSON); err != nil {
		log.Error("[Create Team] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	if teamJSON.TeamName == "" || teamJSON.OrgName == "" {
		log.Error("[Create Team] Team name and organization name can not be empty")
		return http.StatusBadRequest, []byte("Team name and organization name can not be empty")
	}

	//3. get permission
	permisson, err := getOrganizaionPermission(user, teamJSON.OrgName)
	if err != nil {
		log.Error("[Create Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN {
		log.Error("[Create Team] Unauthorized to create team")
		return http.StatusUnauthorized, []byte("Unauthorized to create team")
	}

	//4. save, must be transaction
	team := &dao.Team{
		Name:    teamJSON.TeamName,
		Comment: teamJSON.Comment,
		Org:     &dao.Organization{Name: teamJSON.OrgName},
	}
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: teamJSON.TeamName,
			Org:  &dao.Organization{Name: teamJSON.OrgName},
		},
		User: &dao.User{Name: userName},
		Role: dao.TEAMADMIN,
	}
	if err := dao.CreateTeam(team, tum); err != nil {
		log.Error("[Create Team] Failed to save to db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

func DeleteTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Delete Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Delete Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	if teamName == "" || orgName == "" {
		log.Error("[Delete Team] Team name and organization name can not be empty")
		return http.StatusBadRequest, []byte("Team name and organization name can not be empty")
	}

	//3. get permission
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Delete Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Delete Team] Unauthorized to delete team")
		return http.StatusUnauthorized, []byte("Unauthorized to delete team")
	}

	//4. delete
	team := &dao.Team{
		Name: teamName,
		Org:  &dao.Organization{Name: orgName},
	}
	if err := team.Delete(); err != nil {
		log.Error("[Delete Team]Failed to delete team:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

func DeactiveTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	if teamName == "" || orgName == "" {
		log.Error("[Deactive Team] Team name and organization name can not be empty")
		return http.StatusBadRequest, []byte("Team name and organization name can not be empty")
	}

	//3. get permission
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Deactive Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Deactive Team] Unauthorized to deactive team")
		return http.StatusUnauthorized, []byte("Unauthorized to deactive team")
	}

	//4. deactive
	team := &dao.Team{
		Name: teamName,
		Org:  &dao.Organization{Name: orgName},
	}
	if err := team.Deactive(); err != nil {
		log.Error("[Deactive Team]Failed to deactive team:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

// ActiveTeam returns response of active according to the request
func ActiveTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active Team] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	if teamName == "" || orgName == "" {
		log.Error("[Active Team] Team name and organization name can not be empty")
		return http.StatusBadRequest, []byte("Team name and organization name can not be empty")
	}

	//3. get permission
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Active Team] Failed to get user permission: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Active Team] Unauthorized to active team")
		return http.StatusUnauthorized, []byte("Unauthorized to active team")
	}

	//4. active
	team := &dao.Team{
		Name: teamName,
		Org:  &dao.Organization{Name: orgName},
	}
	if err := team.Active(); err != nil {
		log.Error("[Active Team]Failed to active team: " + err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

func getRepositoryTeamPermission(user *dao.User, orgName, teamName string) (int, error) {
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
	} else if exist && oum.Role == dao.ORGADMIN {
		return oum.Role, nil
	}

	//user is team admin
	tum := &dao.TeamUserMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName},
		},
		User: &dao.User{Name: user.Name},
	}
	if exist, err := tum.Get(); err != nil {
		return 0, err
	} else if !exist {
		return 0, nil
	} else {
		return tum.Role, nil
	}
}

func GetTeamListFromOrganization(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get Team List From Organization] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get Team List From Organization] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get permission
	orgName := ctx.Params(":organization")
	if permisson, err := getOrganizaionPermission(user, orgName); err != nil {
		log.Error("[Get Team List From Organization] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.ORGMEMBER {
		log.Error("[Get Team List From Organization] Unauthorized to get team list from organization")
		return http.StatusUnauthorized, []byte("Unauthorized to get team list from organization")
	}

	//3. get team list from organization
	team := &dao.Team{Org: &dao.Organization{Name: orgName}}
	if teamlist, err := team.List(); err != nil {
		log.Error("[Get Team List From Organization] Failed to get team list from organization: %v", err.Error())
		return ConvertError(err)
	} else {
		teamjsonlist := []TeamJSON{}
		for _, eachteam := range teamlist {
			eachteamjson := TeamJSON{
				TeamName: eachteam.Name,
				OrgName:  orgName,
				Comment:  eachteam.Comment,
				Status:   eachteam.Status,
			}
			teamjsonlist = append(teamjsonlist, eachteamjson)
		}
		if result, err := json.Marshal(teamjsonlist); err != nil {
			log.Error("[Get Team List From Organization] Failed to marshal team list in organization: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}

func UpdateTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get request info
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Team] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//get update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Team] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "teamname" && k != "orgname" && k != "comment" {
			log.Error("[Update Team] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	fields := []string{}
	team := &dao.Team{Org: &dao.Organization{Name: ""}}
	if val, ok := dat["teamname"]; ok {
		team.Name = val.(string)
	}
	if val, ok := dat["orgname"]; ok {
		team.Org = &dao.Organization{Name: val.(string)}
	}
	if val, ok := dat["comment"]; ok {
		team.Comment = val.(string)
		fields = append(fields, "comment")
	}

	if team.Name == "" || team.Org.Name == "" {
		log.Error("[Update Team] Team name and organization name can not be empty")
		return http.StatusBadRequest, []byte("Team name and organization name can not be empty")
	}

	//3. get permisson
	permisson, err := getRepositoryTeamPermission(user, team.Org.Name, team.Name)
	if err != nil {
		log.Error("[Update Team] Failed to get user permission: %v", err.Error())
		return ConvertError(err)
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Update Team] Unauthorized to update team info")
		return http.StatusUnauthorized, []byte("Unauthorized to update team info")
	}

	//4. update
	if err := team.Update(fields...); err != nil {
		log.Error("[Update Team] Failed to update team info to db:%v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}
