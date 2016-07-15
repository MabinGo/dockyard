package teamtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
	"github.com/containerops/dockyard/utils/setting"
)

var tAdmin, tMember, user1, user2 *dao.User
var org *dao.Organization

func Test_TeamInit(t *testing.T) {
	teamtestInit(t)

	//1. create user1
	user1 = &dao.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, user1)

	user2 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, user2)

	tAdmin = &dao.User{
		Name:     "tadmin",
		Email:    "tadmin@gmail.com",
		Password: "tadmin",
		RealName: "tadmin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, tAdmin)

	tMember = &dao.User{
		Name:     "tmember",
		Email:    "tmember@gmail.com",
		Password: "tmember",
		RealName: "tmember",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, tMember)

	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
	}

	//2. create organization
	org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	authtest.CreateOrganization(t, org, user1.Name, user1.Password)

	//add user2 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON, user1.Name, user1.Password)

	//add teamAdmin to organization
	oumJSON1 := &controller.OrganizationUserMapJSON{
		UserName: tAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON1, user1.Name, user1.Password)

	//add teamMember to organization
	oumJSON2 := &controller.OrganizationUserMapJSON{
		UserName: tMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON2, user1.Name, user1.Password)
}

//
//1.============================== Test CreateTeam API ==============================
//
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
	authtest.SignUp(t, User3)

	teamJSON := &controller.TeamJSON{
		TeamName: "hd_team",
		Comment:  "create team",
		OrgName:  org.Name,
	}

	statusCode, err := authtest.CreateTeam(t, teamJSON, User3.Name, User3.Password)
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
		OrgName:  org.Name,
	}

	statusCode, err := authtest.CreateTeam(t, teamJSON, User4.Name, User4.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Error.")
	}
}

//orgMember Create Team
func Test_OrgMemberCreateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "create team",
		OrgName:  org.Name,
	}

	statusCode, err := authtest.CreateTeam(t, teamJSON, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create team Failed.")
	}
}

//orgadmin Create Team
func Test_OrgAdminCreateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "create team",
		OrgName:  org.Name,
	}

	statusCode, err := authtest.CreateTeam(t, teamJSON, user1.Name, user1.Password)
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
func Test_OrgAdminCreateSameTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "create team",
		OrgName:  org.Name,
	}

	statusCode, err := authtest.CreateTeam(t, teamJSON, user1.Name, user1.Password)
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

//
//2.============================== Test UpdateTeam API ==============================
//
//Update Init
func Test_UpateteamInit(t *testing.T) {

	//teamAdmin add to team
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "hw_team",
		OrgName:  org.Name,
		UserName: tAdmin.Name,
		Role:     dao.TEAMADMIN,
	}
	authtest.AddUserToTeam(t, tumJSON, user1.Name, user1.Password)

	//teamMember add to team
	tumJSON1 := &controller.TeamUserMapJSON{
		TeamName: "hw_team",
		OrgName:  org.Name,
		UserName: tMember.Name,
		Role:     dao.TEAMMEMBER,
	}
	authtest.AddUserToTeam(t, tumJSON1, user1.Name, user1.Password)
}

//Update Team
func UpdateTeamTest(t *testing.T, teamJSON *controller.TeamJSON, username, password string) (int, error) {
	body, _ := json.Marshal(teamJSON)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+authtest.Domains+"/uam/team/update", bytes.NewBuffer(body))
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
func Test_OrgAdminUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "orgadmin update team",
		OrgName:  org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, user1.Name, user1.Password)
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
func Test_SysAdminUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "sysAdmin update team",
		OrgName:  org.Name,
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
func Test_TeamAdminUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "teamAdmin update team",
		OrgName:  org.Name,
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
func Test_OrgMemberUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "orgMember update team",
		OrgName:  org.Name,
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update team error.")
	}
}

//teamMember update Team
func Test_TeamMemberUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "teamMember update team",
		OrgName:  org.Name,
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

//sysAdmin Use Wrong Parameter update Team
func Test_UseWrongParameterUpdateTeam(t *testing.T) {

	teamJSON := &controller.TeamJSON{
		TeamName: "hw_team",
		Comment:  "sysAdmin update team",
		OrgName:  org.Name,
		Status:   1, // Status can not update
	}

	statusCode, err := UpdateTeamTest(t, teamJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update team error.")
	}
}

//
//3.============================== Test DeleteTeam API ==============================
//
// Delete Team
func DeleteTeamTest(t *testing.T, OrgName, team, userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+authtest.Domains+"/uam/team/"+OrgName+"/"+team, nil)
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

	statusCode, err := DeleteTeamTest(t, org.Name, "hw_team", User4.Name, User4.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//orgMember delete Team
func Test_OrgMemberDeleteTeam(t *testing.T) {

	statusCode, err := DeleteTeamTest(t, org.Name, "hw_team", user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//orgadmin delete Team
func Test_OrgAdminDeleteTeam(t *testing.T) {

	statusCode, err := DeleteTeamTest(t, org.Name, "hw_team", user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Delete Team Failed.")
	}
}

//
//4.============================== Test DeactiveTeam API ==============================
//
func Test_DeactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeam: %v", err.Error())
	}
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeam Error")
	}
	tearDownTest(t)
}

//
//5.============================== Test ActiveTeam API ==============================
//
func Test_ActiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/active/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/active/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeam: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/active/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeam: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveTeam Error")
	}
	tearDownTest(t)
}

func Test_PushActiveTeam(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/active/" + team1.OrgName + "/" + team1.TeamName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeam: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Errorf("PushActiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

//clear test
func Test_TeamClear(t *testing.T) {
	authtest.DeleteOrganization(t, org.Name, user1.Name, user1.Password)
	user3 := &dao.User{Name: "user3"}
	authtest.DeleteUser(t, user3.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user2.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, tAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, tMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamMember.Name, sysAdmin.Name, sysAdmin.Password)
}
