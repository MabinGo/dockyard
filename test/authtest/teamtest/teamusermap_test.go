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

var orgHw *dao.Organization
var team *dao.Team
var admin, usert1, usert2, user3 *dao.User

func Test_TeamUserMapInit(t *testing.T) {
	teamtestInit(t)

	//1. create orgAdmin user1 user2
	admin = &dao.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, admin)

	usert1 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, usert1)

	usert2 = &dao.User{
		Name:     "user2",
		Email:    "usert2@gmail.com",
		Password: "aaaaa",
		RealName: "usert2",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, usert2)

	user3 = &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "aaaaa",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	authtest.SignUp(t, user3)

	//2. create organization
	orgHw = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	authtest.CreateOrganization(t, orgHw, admin.Name, admin.Password)

	//add usert1 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: usert1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON, admin.Name, admin.Password)

	//3, create team
	team = &dao.Team{
		Name: "hw_team",
		Org:  orgHw,
	}

	teamJSON := &controller.TeamJSON{
		TeamName: team.Name,
		Comment:  "create team",
		OrgName:  orgHw.Name,
	}
	authtest.CreateTeam(t, teamJSON, admin.Name, admin.Password)

}

//
//1.============================== Test AddUserToTeam API ==============================
//
//Non TeamMember Add usert2 To Team
func Test_NonTeamMemberAddUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := authtest.AddUserToTeam(t, tumJSON, usert1.Name, usert1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error.")
	}
}

//orgAdmin Add orgHw usert1 To Team
func Test_OrgAdminAddOrgUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert1.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := authtest.AddUserToTeam(t, tumJSON, admin.Name, admin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Failed")
	}
}

//TeamMember Add usert2 To Team
func Test_TeamMemberAddUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := authtest.AddUserToTeam(t, tumJSON, usert1.Name, usert1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error")
	}
}

//orgAdmin Add Non-orgHw User To Team
func Test_OrgAdminAddNonOrgUserToTeam(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := authtest.AddUserToTeam(t, tumJSON, admin.Name, admin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Team Error")
	}
}

//TeamAdmin Add usert2 To soft_team
func Test_TeamAdminAddUserToTeam(t *testing.T) {

	//1. add usert2 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: usert2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON, admin.Name, admin.Password)

	//2. orgAdmin create team1
	team1 := &dao.Team{
		Name: "soft_team",
		Org:  orgHw,
	}
	teamJSON := &controller.TeamJSON{
		TeamName: team1.Name,
		Comment:  "create SoftTeam",
		OrgName:  orgHw.Name,
	}
	authtest.CreateTeam(t, teamJSON, admin.Name, admin.Password)

	//orgAdmin Add usert1 To SoftTeam ,and set usert1 as SoftTeam's TeamAdmin
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team1.Name,
		OrgName:  orgHw.Name,
		UserName: usert1.Name,
		Role:     dao.TEAMADMIN,
	}
	authtest.AddUserToTeam(t, tumJSON, admin.Name, admin.Password)

	//TeamAdmin usert1 Add usert2 To SoftTeam
	tumJSON2 := &controller.TeamUserMapJSON{
		TeamName: team1.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}
	statusCode, err := authtest.AddUserToTeam(t, tumJSON2, usert1.Name, usert1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Failed")
	}
}

//sysAdmin Add usert2 To hw_team
func Test_SysAdminAddOrgUserToTeam(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := authtest.AddUserToTeam(t, tumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Add User To Team Error")
	}
}

//
//2.============================== Test UpdateTeamUserMap API ==============================
//
func UpdateTeamUserMap(t *testing.T, tumJSON *controller.TeamUserMapJSON, username, password string) (int, error) {
	body, _ := json.Marshal(tumJSON)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+authtest.Domains+"/uam/team/updateteamusermap", bytes.NewBuffer(body))
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

//sysAdmin Update TeamUserMap
func Test_SysAdminUpdateTeamUserMap(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMADMIN,
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update TeamUserMap Failed")
	}
}

//orgAdmin Update TeamUserMap
func Test_OrgAdminUpdateTeamUserMap(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMMEMBER,
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, admin.Name, admin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update TeamUserMap Failed")
	}
}

//orgmember(soft_team) Update TeamUserMap
func Test_OrgMemberUpdateTeamUserMap(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "soft_team",
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMADMIN,
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, user3.Name, user3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update TeamUserMap Error")
	}
}

//teamMember(soft_team) Update TeamUserMap
func Test_TeamMemberUpdateTeamUserMap(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "soft_team",
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMADMIN,
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, usert2.Name, usert2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update TeamUserMap Error")
	}
}

//teamAdmin(soft_team) Update TeamUserMap
func Test_TeamAdminUpdateTeamUserMap(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "soft_team",
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMADMIN,
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, usert1.Name, usert1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update TeamUserMap Failed")
	}

	//if not update back it will affect below test func
	tumJSON.Role = dao.TEAMMEMBER
	UpdateTeamUserMap(t, tumJSON, usert1.Name, usert1.Password)
}

//sysAdmin Use Wrong Parameter Update TeamUserMap
func Test_UseWrongParameterUpdateTeamUserMap(t *testing.T) {

	tumJSON := &controller.TeamUserMapJSON{
		TeamName: team.Name,
		OrgName:  orgHw.Name,
		UserName: usert2.Name,
		Role:     dao.TEAMADMIN,
		Status:   1, // Status can not update
	}

	statusCode, err := UpdateTeamUserMap(t, tumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update TeamUserMap Error")
	}
}

//
//3.============================== Test RemoveUserFromTeam API ==============================
//
func RemoveUserFromTeam(t *testing.T, orgName, teamName, teamMember, username, password string) (int, error) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+authtest.Domains+"/uam/team/removeuser/"+orgName+"/"+teamName+"/"+teamMember, nil)
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

//sysAdmin Remove usert2 From Team
func Test_SysAdminRemoveUserFromTeam(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, team.Name, usert2.Name, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}

}

//orgAdmin Remove usert1 From Team
func Test_OrgAdminRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, team.Name, usert1.Name, admin.Name, admin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}

}

//TeamMember Remove usert2 From Team
func Test_TeamMemberRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, "soft_team", usert2.Name, usert2.Name, usert2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}
}

//TeamAdmin Remove usert2 From Team
func Test_TeamAdminRemoveUserFromTeam(t *testing.T) {

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, "soft_team", usert2.Name, usert1.Name, usert1.Password)
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

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, "soft_team", usert1.Name, usert2.Name, usert2.Password)
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

	//1. add user3 to organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user3.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}
	authtest.AddUserToOrganization(t, oumJSON, admin.Name, admin.Password)

	//orgAdmin Add user3 To SoftTeam
	tumJSON := &controller.TeamUserMapJSON{
		TeamName: "soft_team",
		OrgName:  orgHw.Name,
		UserName: user3.Name,
		Role:     dao.TEAMMEMBER,
	}
	authtest.AddUserToTeam(t, tumJSON, admin.Name, admin.Password)

	statusCode, err := RemoveUserFromTeam(t, orgHw.Name, "soft_team", user3.Name, usert2.Name, usert2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Log(statusCode)
		t.Fatal("Remove User From Team Failed")
	}
}

//
//4.============================== Test DeactiveTeamUserMap API ==============================
//
func Test_DeactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamUserMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeamUserMap Error")
	}
	tearDownTest(t)
}

//
//5.============================== Test ActiveTeamUserMap API ==============================
//
func Test_ActiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activeuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activeuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activeuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeamUserMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveTeamUserMap Error")
	}
	tearDownTest(t)
}

//clear test
func Test_TeamUserMapClear(t *testing.T) {
	authtest.DeleteOrganization(t, orgHw.Name, admin.Name, admin.Password)
	authtest.DeleteUser(t, usert1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, usert2.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user3.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, admin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamMember.Name, sysAdmin.Name, sysAdmin.Password)
}
