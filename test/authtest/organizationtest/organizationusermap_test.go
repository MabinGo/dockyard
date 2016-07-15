package organizationtest

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

func Test_OrganizationUserMapInit(t *testing.T) {
	organizationtestInit(t)
	authtest.CreateOrganization(t, orgHw, orgAdmin.Name, orgAdmin.Password)
}

//
//1.============================== Test AddUserToOrganization API ==============================
//
//orgAdmin add user1 to Organization
func Test_AdminAddUserToOrganization(t *testing.T) {

	//add u1 to org
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := authtest.AddUserToOrganization(t, oumJSON, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Add User To Organization Failed.")

	}
}

//user2 add user2 to Organization
func Test_UserAddHimselfToOrganization(t *testing.T) {

	//add user2 to org
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := authtest.AddUserToOrganization(t, oumJSON, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

// orgMember add user2 to Organization
func Test_OrgMemberAddUserToOrganization(t *testing.T) {

	//user1 add user2 to org
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := authtest.AddUserToOrganization(t, oumJSON, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

//orgAdmin add Non-existed user to Organization
func Test_AdminAddNonExistedUserToOrganization(t *testing.T) {

	//add user3 to org
	user3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user3.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := authtest.AddUserToOrganization(t, oumJSON, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

//
//2.============================== Test UpdateOrganizationUserMap API ==============================
//
func UpdateOrganizationUserMapTest(t *testing.T, oumJSON *controller.OrganizationUserMapJSON,
	userName, password string) (int, error) {

	body, _ := json.Marshal(oumJSON)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+authtest.Domains+"/uam/organization/updateorganizationusermap",
		bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(userName, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

//orgAdmin Update OrganizationUserMap
func Test_OrgAdminUpdateOrganizationUserMap(t *testing.T) {
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGADMIN,
		OrgName:  orgHw.Name,
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update OrganizationUserMap Failed.")
	}
}

//sysAdmin Update OrganizationUserMap
func Test_SysAdminUpdateOrganizationUserMap(t *testing.T) {
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update OrganizationUserMap Failed.")
	}
}

//sysAdmin Update non exist OrganizationUserMap
func Test_SysAdminUpdateNonExistOrganizationUserMap(t *testing.T) {
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: sysAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  "empty_oum",
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update OrganizationUserMap Failed.")
	}
}

// orgMember (user1) Update OrganizationUserMap
func Test_OrgMemberUpdateOrganizationUserMap(t *testing.T) {

	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGADMIN,
		OrgName:  orgHw.Name,
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update OrganizationUserMap Error.")
	}
}

func Test_UseWrongParameterUpdateOrgUserMap(t *testing.T) {

	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
		Status:   1, // Status can not update
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update OrganizationUserMap Error.")
	}
}

//
//3.============================== Test RemoveUserFromOrganization API ==============================
//
func Test_RemoveUserInit(t *testing.T) {
	//add user2 to Organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  orgHw.Name,
	}

	statusCode, err := authtest.AddUserToOrganization(t, oumJSON, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Add User To Organization Failed.")
	}
}

func RemoveUserFromOrganizationTest(t *testing.T, orgName, delUserName,
	userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+authtest.Domains+"/uam/organization/removeuser/"+orgName+"/"+delUserName, nil)
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

// user1 delete user2
func Test_OrgMemberRemoveUserFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, orgHw.Name, user2.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}

/* // user2 delete user2
func Test_UserRemoveHimselfFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, orgHw.Name, user2.Name, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}*/

// orgAdmin delete user1
func Test_AdminRemoveUserFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, orgHw.Name, user1.Name, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}

func Test_AdminRemoveNonExistedUserFromOrganization(t *testing.T) {

	//delete user3
	user3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	statusCode, err := RemoveUserFromOrganizationTest(t, orgHw.Name, user3.Name, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}

//
//4.============================== Test DeactiveOrgUserMap API ==============================
//
func Test_DeactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveOrgUserMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveOrgUserMap Error")
	}
	tearDownTest(t)
}

//
//5.============================== Test ActiveOrgUserMap API ==============================
//
func Test_ActiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/activeuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrgUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/activeuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveOrgUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveOrgUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/activeuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrgUserMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveOrgUserMap Error")
	}
	tearDownTest(t)
}

func Test_PushActiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/activeuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrgUserMap: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

func Test_PullActiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/activeuser/" + org1.Name + "/" + orgMember.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrgUserMap: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

//clear test
func Test_OrganizationUserMapClear(t *testing.T) {

	authtest.DeleteOrganization(t, orgHw.Name, orgAdmin.Name, orgAdmin.Password)
	authtest.DeleteUser(t, user1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user2.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
}
