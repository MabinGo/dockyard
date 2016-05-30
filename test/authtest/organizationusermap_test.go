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

var OrgAdmin, user1, user2 *dao.User

func Test_organizationusermapInit(t *testing.T) {

	//1.add OrgAdmin, user1, user2
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

	user1 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(user1, t)

	user2 = &dao.User{
		Name:     "user",
		Email:    "user@gmail.com",
		Password: "user",
		RealName: "user",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	signUp(user2, t)

	//2.create organization
	Org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	CreateOrganizationTest(t, Org, OrgAdmin.Name, OrgAdmin.Password)
}

func AddUserToOrganizationTest(t *testing.T, oumJSON *controller.OrganizationUserMapJSON,
	userName, password string) (int, error) {

	body, _ := json.Marshal(oumJSON)
	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/organization/adduser",
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

//OrgAdmin add user1 to Organization
func Test_AdminAddUserToOrganization(t *testing.T) {

	//add u1 to org
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}

	statusCode, err := AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)
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
		OrgName:  Org.Name,
	}

	statusCode, err := AddUserToOrganizationTest(t, oumJSON, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

// orgMember add user2 to Organization
func Test_orgMemberAddUserToOrganization(t *testing.T) {

	//user1 add user2 to org
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}

	statusCode, err := AddUserToOrganizationTest(t, oumJSON, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

//OrgAdmin add Non-existed user to Organization
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
		OrgName:  Org.Name,
	}

	statusCode, err := AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Add User To Organization Error.")
	}
}

func UpdateOrganizationUserMapTest(t *testing.T, oumJSON *controller.OrganizationUserMapJSON,
	userName, password string) (int, error) {

	body, _ := json.Marshal(oumJSON)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+Domains+"/uam/organization/updateorganizationusermap",
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

//OrgAdmin Update OrganizationUserMap
func Test_OrgAdminUpdateOrganizationUserMap(t *testing.T) {
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGADMIN,
		OrgName:  Org.Name,
	}

	statusCode, err := UpdateOrganizationUserMapTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update OrganizationUserMap Failed.")
	}
}

//sysAdmin Update OrganizationUserMap
func Test_sysAdminUpdateOrganizationUserMap(t *testing.T) {
	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
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
func Test_sysAdminUpdateNonExistOrganizationUserMap(t *testing.T) {
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
func Test_orgMemberUpdateOrganizationUserMap(t *testing.T) {

	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user1.Name,
		Role:     dao.ORGADMIN,
		OrgName:  Org.Name,
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

func Test_RemoveUserInit(t *testing.T) {
	//add user2 to Organization
	oumJSON := &controller.OrganizationUserMapJSON{
		UserName: user2.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  Org.Name,
	}

	statusCode, err := AddUserToOrganizationTest(t, oumJSON, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Add User To Organization Failed.")
	}
}

func RemoveUserFromOrganizationTest(t *testing.T, orgName, delUserName,
	userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/organization/removeuser/"+orgName+"/"+delUserName, nil)
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
func Test_orgMemberRemoveUserFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, Org.Name, user2.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}

/* // user2 delete user2
func Test_UserRemoveHimselfFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, Org.Name, user2.Name, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}*/

// OrgAdmin delete user1
func Test_AdminRemoveUserFromOrganization(t *testing.T) {

	statusCode, err := RemoveUserFromOrganizationTest(t, Org.Name, user1.Name, OrgAdmin.Name, OrgAdmin.Password)
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

	statusCode, err := RemoveUserFromOrganizationTest(t, Org.Name, user3.Name, OrgAdmin.Name, OrgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode == 200 {
		t.Fatal("Remove User From Organization Error.")
	}
}

//clear test
func Test_organizationusermapClear(t *testing.T) {

	DeleteOrganizationTest(t, Org.Name, OrgAdmin.Name, OrgAdmin.Password)
	if err := user1.Delete(); err != nil {
		t.Error(err)
	}
	if err := user2.Delete(); err != nil {
		t.Error(err)
	}
	if err := OrgAdmin.Delete(); err != nil {
		t.Error(err)
	}
}
