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

var tAdmin, tMember *dao.User

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

	tAdmin = &dao.User{
		Name:     "tadmin",
		Email:    "tadmin@gmail.com",
		Password: "tadmin",
		RealName: "tadmin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(tAdmin, t)

	tMember = &dao.User{
		Name:     "tmember",
		Email:    "tmember@gmail.com",
		Password: "tmember",
		RealName: "tmember",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(tMember, t)

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

	//add teamAdmin to organization
	oumJSON1 := &controller.OrganizationUserMapJSON{
		UserName: tAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON1, User1.Name, User1.Password)

	//add teamMember to organization
	oumJSON2 := &controller.OrganizationUserMapJSON{
		UserName: tMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON2, User1.Name, User1.Password)
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
		TeamName: "hd_team",
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
		TeamName: "hw_team",
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
		TeamName: "hw_team",
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
		TeamName: "hw_team",
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
		TeamName: "hw_team",
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

//Update
func Test_UpateteamInit(t *testing.T) {

	//teamAdmin add to team
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "hw_team",
		OrgName:  Org.Name,
		UserName: tAdmin.Name,
		Role:     dao.TEAMADMIN,
	}
	AddUserToTeam(t, tumJSON, OrgAdmin.Name, OrgAdmin.Password)

	//teamMember add to team
	tumJSON1 := &controller.TeamUserMapJSON{
		TeamName: "hw_team",
		OrgName:  Org.Name,
		UserName: tMember.Name,
		Role:     dao.TEAMMEMBER,
	}
	AddUserToTeam(t, tumJSON1, OrgAdmin.Name, OrgAdmin.Password)
}

func UpdateTeamTest(t *testing.T, teamJSON *controller.TeamJSON, username, password string) (int, error) {
	body, _ := json.Marshal(teamJSON)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+Domains+"/uam/team/update", bytes.NewBuffer(body))
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

//orgadmin update Team
func Test_orgAdminUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "orgadmin update team",
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update team Failed.")
	}

	// query Team
	team1 := &dao.Team{
		Name: teamJSON.TeamName,
		Org:  &dao.Organization{Name: teamJSON.OrgName}}
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

//sysAdmin update Team
func Test_sysAdminUpdateTeam(t *testing.T) {

	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
	}

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "sysAdmin update team",
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update team Failed.")
	}
}

//teamAdmin update Team
func Test_teamAdminUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "teamAdmin update team",
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, tAdmin.Name, tAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update team Failed.")
	}
}

//orgMember update Team
func Test_orgMemberUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "orgMember update team",
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update team error.")
	}
}

//teamMember update Team
func Test_teamMemberUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "teamMember update team",
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, tMember.Name, tMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update team error.")
	}
}

// Delete
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

	statusCode, err := DeleteTeamTest(t, Org.Name, "hw_team", User4.Name, User4.Password)
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

	statusCode, err := DeleteTeamTest(t, Org.Name, "hw_team", User2.Name, User2.Password)
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

	statusCode, err := DeleteTeamTest(t, Org.Name, "hw_team", User1.Name, User1.Password)
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
	if err := tAdmin.Delete(); err != nil {
		t.Error(err)
	}
	if err := tMember.Delete(); err != nil {
		t.Error(err)
	}
}
