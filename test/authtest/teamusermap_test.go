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

var teams *dao.Team
var User3 *dao.User

func Test_teamusermapInit(t *testing.T) {

	//1. create OrgAdmin user1 user2
	OrgAdmin = &dao.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(OrgAdmin, t)

	User1 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User1, t)

	User2 = &dao.User{
		Name:     "User2",
		Email:    "User2@gmail.com",
		Password: "aaaaa",
		RealName: "User2",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User2, t)

	User3 = &dao.User{
		Name:     "User3",
		Email:    "User3@gmail.com",
		Password: "aaaaa",
		RealName: "User3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(User3, t)

	//2. create organization
	Org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	CreateOrganizationTest(t, Org, OrgAdmin.Name, OrgAdmin.Password)

	//add User1 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: User1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)

	//3, create team
	teams = &dao.Team{
		Name: "HWTeam",
		Org:  Org,
	}

	teamJSON := &controller.TeamJSON{
		TeamName: teams.Name,
		Comment:  "create team",
		OrgName:  Org.Name,
	}
	CreateTeamTest(t, teamJSON, OrgAdmin.Name, OrgAdmin.Password)

}

func AddUserToTeam(t *testing.T, tumJSON *controller.TeamUserMapJSON, username, password string) (int, error) {
	body, _ := json.Marshal(tumJSON)
	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/team/adduser", bytes.NewBuffer(body))
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

//Non TeamMember Add User2 To Team
func Test_NonTeamMemberAddUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: teams.Name,
		OrgName:  Org.Name,
		UserName: User2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := AddUserToTeam(t, tumJSON, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error.")
	}
}

//OrgAdmin Add Org User1 To Team
func Test_OrgAdminAddOrgUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: teams.Name,
		OrgName:  Org.Name,
		UserName: User1.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := AddUserToTeam(t, tumJSON, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Failed")
	}
}

//TeamMember Add User2 To Team
func Test_TeamMemberAddUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: teams.Name,
		OrgName:  Org.Name,
		UserName: User2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := AddUserToTeam(t, tumJSON, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error")
	}
}

//OrgAdmin Add Non-Org User To Team
func Test_OrgAdminAddNonOrgUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: teams.Name,
		OrgName:  Org.Name,
		UserName: User2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := AddUserToTeam(t, tumJSON, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error")
	}
}

//TeamAdmin Add User2 To Team
func Test_TeamAdminAddUserToTeam(t *testing.T) {

	//1. add User2 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: User2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)

	//2. OrgAdmin create team1
	team1 := &dao.Team{
		Name: "SoftTeam",
		Org:  Org,
	}
	teamJSON := &controller.TeamJSON{
		TeamName: team1.Name,
		Comment:  "create SoftTeam",
		OrgName:  Org.Name,
	}
	CreateTeamTest(t, teamJSON, OrgAdmin.Name, OrgAdmin.Password)

	//OrgAdmin Add User1 To SoftTeam ,and set User1 as SoftTeam's TeamAdmin
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team1.Name,
		OrgName:  Org.Name,
		UserName: User1.Name,
		Role:     dao.TEAMADMIN,
	}
	AddUserToTeam(t, tumJSON, OrgAdmin.Name, OrgAdmin.Password)

	//TeamAdmin User1 Add User2 To SoftTeam
	tumJSON2 := &controller.TeamUserMapJSON{
		TeamName: team1.Name,
		OrgName:  Org.Name,
		UserName: User2.Name,
		Role:     dao.TEAMMEMBER,
	}
	statusCode, err := AddUserToTeam(t, tumJSON2, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Failed")
	}
}

//sysAdmin Add User To Team
func Test_sysAdminAddOrgUserToTeam(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: teams.Name,
		OrgName:  Org.Name,
		UserName: User2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := AddUserToTeam(t, tumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Error")
	}
}

func RemoveUserFromTeam(t *testing.T, orgName, teamName, teamMember, username, password string) (int, error) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/team/removeuser/"+orgName+"/"+teamName+"/"+teamMember, nil)
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

//sysAdmin Remove User2 From Team
func Test_sysAdminRemoveUserFromTeam(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	statusCode, err := RemoveUserFromTeam(t, Org.Name, teams.Name, User2.Name, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}

}

//OrgAdmin Remove User1 From Team
func Test_OrgAdminRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, Org.Name, teams.Name, User1.Name, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}

}

//TeamMember Remove User2 From Team
func Test_TeamMemberRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, Org.Name, "SoftTeam", User2.Name, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}
}

//TeamAdmin Remove User2 From Team
func Test_TeamAdminRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, Org.Name, "SoftTeam", User2.Name, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}

}

//Non TeamMember Remove TeamAdmin From Team
func Test_NonTeamMemberRemoveTeamAdminFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, Org.Name, "SoftTeam", User1.Name, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}
}

func Test_NonTeamMemberRemoveUserFromTeam(t *testing.T) {

	//1. add User3 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: User3.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)

	//OrgAdmin Add User3 To SoftTeam
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "SoftTeam",
		OrgName:  Org.Name,
		UserName: User3.Name,
		Role:     dao.TEAMMEMBER,
	}
	AddUserToTeam(t, tumJSON, OrgAdmin.Name, OrgAdmin.Password)

	statusCode, err := RemoveUserFromTeam(t, Org.Name, "SoftTeam", User3.Name, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}
}

//clear test
func Test_teamusermapClear(t *testing.T) {

	DeleteOrganizationTest(t, Org.Name, OrgAdmin.Name, OrgAdmin.Password)
	if err := User1.Delete(); err != nil {
		t.Error(err)
	}
	if err := User2.Delete(); err != nil {
		t.Error(err)
	}
	if err := User3.Delete(); err != nil {
		t.Error(err)
	}
	if err := OrgAdmin.Delete(); err != nil {
		t.Error(err)
	}
}
