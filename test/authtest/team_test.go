package authtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func Test_teamInit(t *testing.T) {

	//1. create user1
	User1 = &dao.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User1, t)

	User2 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User2, t)

	//2. create organization
	Org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	CreateOrganizationTest(t, Org, User1.Name, User1.Password)

	//add User2 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: User2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, User1.Name, User1.Password)
}

func CreateTeamTest(t *testing.T, teamJSON *controller.TeamJSON, username, password string) (int, error) {
	body, _ := json.Marshal(teamJSON)
	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/team", bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(username, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

//Non orgMember Create Team
func Test_NonOrgMemberCreateTeam(t *testing.T) {

	User3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User3, t)

	teamJSON := &controller.TeamJSON{
		TeamName: "HDTeam",
		Comment:  "create team",
		OrgName:  Org.Name,
	}

	statusCode, err := CreateTeamTest(t, teamJSON, User3.Name, User3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Error.")
	}
}

func Test_NonExsitedUserCreateTeam(t *testing.T) {
	User4 := &dao.User{
		Name:     "user4",
		Email:    "user4@gmail.com",
		Password: "user4",
		RealName: "user4",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	teamJSON := &controller.TeamJSON{
		TeamName: "HWTeam",
		Comment:  "create team",
		OrgName:  Org.Name,
	}

	statusCode, err := CreateTeamTest(t, teamJSON, User4.Name, User4.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Error.")
	}
}

//orgMember Create Team
func Test_orgMemberCreateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "HWTeam",
		Comment:  "create team",
		OrgName:  Org.Name,
	}

	statusCode, err := CreateTeamTest(t, teamJSON, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Failed.")
	}
}

//orgadmin Create Team
func Test_orgAdminCreateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "HWTeam",
		Comment:  "create team",
		OrgName:  Org.Name,
	}

	statusCode, err := CreateTeamTest(t, teamJSON, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create team Failed.")
	}

	// query Team
	team1 := &dao.Team{Name: teamJSON.TeamName,
		Org: &dao.Organization{Name: teamJSON.OrgName}}
	if exist, err := team1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("team is not exitst")
	} else {
		if team1.Name != teamJSON.TeamName || team1.Org.Name != teamJSON.OrgName ||
			team1.Comment != teamJSON.Comment {
			t.Error("team1's save is not same with get")
		}
	}
}

//orgadmin re Create Team
func Test_orgAdminCreateSameTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "HWTeam",
		Comment:  "create team",
		OrgName:  Org.Name,
	}

	statusCode, err := CreateTeamTest(t, teamJSON, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Failed.")
	}

	// query Team
	team1 := &dao.Team{Name: teamJSON.TeamName,
		Org: &dao.Organization{Name: teamJSON.OrgName}}
	if exist, err := team1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("team is not exitst")
	} else {
		if team1.Name != teamJSON.TeamName || team1.Org.Name != teamJSON.OrgName ||
			team1.Comment != teamJSON.Comment {
			t.Error("team1's save is not same with get")
		}
	}
}

func DeleteTeamTest(t *testing.T, OrgName, team, userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/team/"+OrgName+"/"+team, nil)
	if err != nil {
		return -1, err
	}
	req.SetBasicAuth(userName, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

//Non Exsied user delete Team
func Test_NonExsieduserDeleteTeam(t *testing.T) {

	User4 := &dao.User{
		Name:     "user4",
		Email:    "user4@gmail.com",
		Password: "user4",
		RealName: "user4",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	statusCode, err := DeleteTeamTest(t, Org.Name, "HWTeam", User4.Name, User4.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//orgMember delete Team
func Test_orgMemberDeleteTeam(t *testing.T) {

	statusCode, err := DeleteTeamTest(t, Org.Name, "HWTeam", User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//orgadmin delete Team
func Test_orgAdminDeleteTeam(t *testing.T) {

	statusCode, err := DeleteTeamTest(t, Org.Name, "HWTeam", User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//clear test
func Test_TeamTestClear(t *testing.T) {

	DeleteOrganizationTest(t, Org.Name, User1.Name, User1.Password)
	if err := User1.Delete(); err != nil {
		t.Error(err)
	}
	if err := User2.Delete(); err != nil {
		t.Error(err)
	}
	User3 := &dao.User{Name: "user3"}
	if err := User3.Delete(); err != nil {
		t.Error(err)
	}
}
