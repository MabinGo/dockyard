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

type TeamRepositoryMapJSON struct {
	OrgName  string `json:"orgname"`
	RepoName string `json:"reponame"`
	TeamName string `json:"teamname"`
	Permit   int    `json:"permit"`           //team permit for repo ,WRITE or READ
	Status   int    `json:"status,omitempty"` //status: active(0) or inactive(1)
}

//input:orgName,teamName,repoName, permit
//login user must be system admin, org admin or  team  admin.
//team and repositoy must be in orgnazaiton.
func AddRepositoryToTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn for user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Add Repository To Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Add Repository To Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgname, teamName, repoName, permit
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Add Repository To Team] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	trmJSON := &TeamRepositoryMapJSON{}
	if err := json.Unmarshal(body, trmJSON); err != nil {
		log.Error("[Add Repository To Team] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//3. get permisson for adding repo in team
	permisson, err := getRepositoryTeamPermission(user, trmJSON.OrgName, trmJSON.TeamName)
	if err != nil {
		log.Error("[Add Repository To Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Add Repository To Team] Unauthorized to add repository to team")
		return http.StatusUnauthorized, []byte("Unauthorized to add repository to team")
	}

	//4. add repo into team
	trm := &dao.TeamRepositoryMap{
		Team: &dao.Team{
			Name: trmJSON.TeamName,
			Org:  &dao.Organization{Name: trmJSON.OrgName},
		},
		Repo: &dao.RepositoryEx{
			Name:     trmJSON.RepoName,
			IsOrgRep: true,
			Org:      &dao.Organization{Name: trmJSON.OrgName},
		},
		Permit: trmJSON.Permit, //team's access permit for repository，
	}
	if err := trm.Save(); err != nil {
		log.Error("[Add Repository To Team] Save data error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, nil
}

//input:orgName,teamName,repoName
//login user must be system admin, org admin or  team  admin.
//team and repositoy must be in orgnazaiton.
func RemoveRepositoryFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn for user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Remove Repository From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Remove Repository From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgname, teamName, repoName
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	repoName := ctx.Params(":repository")

	//3. get permisson for removing repo from team
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Remove Repository From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Remove Repository From Team] Unauthorized to add repository to team")
		return http.StatusUnauthorized, []byte("Unauthorized to add repository to team")
	}

	//4. remove repo from team
	trm := &dao.TeamRepositoryMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName},
		},
		Repo: &dao.RepositoryEx{
			Name:     repoName,
			IsOrgRep: true,
			Org:      &dao.Organization{Name: orgName},
		},
	}
	if err := trm.Delete(); err != nil {
		log.Error("[Remove Repository From Team] Delete data error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, nil
}

//input:orgName,teamName,repoName, permit
//login user must be system admin, org admin or  team  admin.
//team and repositoy must be in orgnazaiton.
func UpdateTeamRepositoryMap(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn for user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update Team Repository Map] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update Team Repository Map] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgname, teamName, repoName, permit
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update Team Repository Map] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}
	trmJSON := &TeamRepositoryMapJSON{}
	if err := json.Unmarshal(body, trmJSON); err != nil {
		log.Error("[Update Team Repository Map] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte(err.Error())
	}

	//check update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update Team Repository Map] Failed to unmarshal request body")
		return http.StatusBadRequest, []byte(err.Error())
	}
	for k := range dat {
		if k != "orgname" && k != "reponame" && k != "teamname" && k != "permit" {
			log.Error("[Update Team Repository Map] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	//3. get permisson for update repo in team
	permisson, err := getRepositoryTeamPermission(user, trmJSON.OrgName, trmJSON.TeamName)
	if err != nil {
		log.Error("[Update Team Repository Map] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Update Team Repository Map] Unauthorized to update team repository map info")
		return http.StatusUnauthorized, []byte("Unauthorized to update team repository map info")
	}

	//4. update team repository info
	trm := &dao.TeamRepositoryMap{
		Team: &dao.Team{
			Name: trmJSON.TeamName,
			Org:  &dao.Organization{Name: trmJSON.OrgName},
		},
		Repo: &dao.RepositoryEx{
			Name:     trmJSON.RepoName,
			IsOrgRep: true,
			Org:      &dao.Organization{Name: trmJSON.OrgName},
		},
		Permit: trmJSON.Permit, //team's access permit for repository，
	}
	if err := trm.Update("Permit"); err != nil {
		log.Error("[Update Team Repository Map] Update team repository map info error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, nil
}

//input:orgName,teamName,repoName
//login user must be system admin, org admin or  team  admin.
//team and repositoy must be in orgnazaiton.
func DeactiveRepositoryFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn for user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive Repository From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive Repository From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgname, teamName, repoName
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	repoName := ctx.Params(":repository")

	//3. get permisson for removing repo from team
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Deactive Repository From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Deactive Repository From Team] Unauthorized to deactive repository to team")
		return http.StatusUnauthorized, []byte("Unauthorized to deactive repository to team")
	}

	//4. deactive repo from team
	trm := &dao.TeamRepositoryMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName},
		},
		Repo: &dao.RepositoryEx{
			Name:     repoName,
			IsOrgRep: true,
			Org:      &dao.Organization{Name: orgName},
		},
	}
	if err := trm.Deactive(); err != nil {
		log.Error("[Deactive Repository From Team] Deactive data error:%v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, nil
}

// ActiveRepositoryFromTeam returns response of active according to the request
//login user must be system admin, org admin or team admin.
//team and repositoy must be in orgnazaiton.
func ActiveRepositoryFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn for user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Add Repository To Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active Repository From Team] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	//2. get orgname, teamName, repoName
	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")
	repoName := ctx.Params(":repository")

	//3. get permisson for removing repo from team
	permisson, err := getRepositoryTeamPermission(user, orgName, teamName)
	if err != nil {
		log.Error("[Active Repository From Team] Failed to get user permission: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN {
		log.Error("[Active Repository From Team] Unauthorized to active repository to team")
		return http.StatusUnauthorized, []byte("Unauthorized to active repository to team")
	}

	//4. active repo from team
	trm := &dao.TeamRepositoryMap{
		Team: &dao.Team{
			Name: teamName,
			Org:  &dao.Organization{Name: orgName},
		},
		Repo: &dao.RepositoryEx{
			Name:     repoName,
			IsOrgRep: true,
			Org:      &dao.Organization{Name: orgName},
		},
	}
	if err := trm.Active(); err != nil {
		log.Error("[Active Repository From Team] Active data error:" + err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, nil
}

//Get: repository list from team
func GetRepositoryListFromTeam(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//1. authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get Repository List From Team] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get Repository List From Team] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	orgName := ctx.Params(":organization")
	teamName := ctx.Params(":team")

	//2. get permisson
	if permisson, err := getRepositoryTeamPermission(user, orgName, teamName); err != nil {
		log.Error("[Get Repository List From Team] Failed to get user permission: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else if permisson != dao.SYSADMIN && permisson != dao.ORGADMIN && permisson != dao.TEAMADMIN && permisson != dao.TEAMMEMBER {
		log.Error("[Get Repository List From Team] Unauthorized to get repository list from team")
		return http.StatusUnauthorized, []byte("Unauthorized to get repository list from team")
	}

	//3. get repository list from team
	trm := &dao.TeamRepositoryMap{Team: &dao.Team{Name: teamName, Org: &dao.Organization{Name: orgName}}}
	if trmlist, err := trm.List(); err != nil {
		log.Error("[Get Repository List From Team] Failed to get repository list from team: %v", err.Error())
		return ConvertError(err)
	} else {
		trmjsonlist := []TeamRepositoryMapJSON{}
		for _, eachtrm := range trmlist {
			eachtrmjson := TeamRepositoryMapJSON{
				OrgName:  orgName,
				RepoName: eachtrm.Repo.Name,
				Permit:   eachtrm.Permit,
				TeamName: teamName,
				Status:   eachtrm.Status,
			}
			trmjsonlist = append(trmjsonlist, eachtrmjson)
		}
		if result, err := json.Marshal(trmjsonlist); err != nil {
			log.Error("[Get Repository List From Team] Failed to marshal repository list in Team: %v", err.Error())
			return http.StatusInternalServerError, []byte(err.Error())
		} else {
			return http.StatusOK, result
		}
	}
}
