package authtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
	"github.com/containerops/dockyard/utils/setting"
)

var User1, User2 *dao.User
var Org *dao.Organization
var openDBFlag bool = false

func openDB(t *testing.T) {

	if openDBFlag {
		return
	}
	if err := db.RegisterDriver("mysql"); err != nil {
		t.Error(err)
	} else {
		db.Drv.RegisterModel(new(dao.User), new(dao.Organization),
			new(dao.RepositoryEx), new(dao.OrganizationUserMap),
			new(dao.Team), new(dao.TeamUserMap), new(dao.TeamRepositoryMap))
		err := db.Drv.InitDB("mysql", setting.DBUser, setting.DBPasswd, setting.DBURI, setting.DBName, 0)
		if err != nil {
			t.Error(err)
		}
	}
	openDBFlag = true
}

func Test_organizationInit(t *testing.T) {
	openDB(t)

	//1. create user1 user2
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

	//2. Init organization struct
	Org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
}

func CreateOrganizationTest(t *testing.T, org *dao.Organization, username, password string) (int, error) {
	body, _ := json.Marshal(org)
	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/organization", bytes.NewBuffer(body))
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

//user1 Create Organization
func Test_CreateOrganization(t *testing.T) {

	statusCode, err := CreateOrganizationTest(t, Org, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create Organization Failed.")
	}

	// query organization
	org1 := &dao.Organization{Name: Org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if org1.Name != Org.Name || org1.Email != Org.Email ||
			org1.Comment != Org.Comment || org1.URL != Org.URL ||
			org1.Location != Org.Location ||
			org1.MemberPrivilege != Org.MemberPrivilege {
			t.Error("org's save is not same with get")
		}
	}
}

//user2 create same Organization with user1
func Test_ReCreateOrganization(t *testing.T) {

	statusCode, err := CreateOrganizationTest(t, Org, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create Organization Error.")
	}
}

//Non-Existed user3 create Organization
func Test_NonExistedUserCreateOrganization(t *testing.T) {

	user3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	// Init organization struct
	org2 := &dao.Organization{
		Name:            "huawei.com",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}

	statusCode, err := CreateOrganizationTest(t, org2, user3.Name, user3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create Organization Error.")
	}

}

// update
func UpdateOrganizationTest(t *testing.T, org *dao.Organization, username, password string) (int, error) {
	body, _ := json.Marshal(org)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+Domains+"/uam/organization/update", bytes.NewBuffer(body))
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

//orgAdmin (user1) update Organization
func Test_orgAdminUpdateOrganization(t *testing.T) {

	Org.MemberPrivilege = dao.READ
	Org.Comment = "orgAdmin update"
	Org.URL = "url test"
	Org.Location = "land"
	statusCode, err := UpdateOrganizationTest(t, Org, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Failed.")
	}

	// query organization
	org1 := &dao.Organization{Name: Org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if org1.Name != Org.Name || org1.Email != Org.Email ||
			org1.Comment != Org.Comment || org1.URL != Org.URL ||
			org1.Location != Org.Location ||
			org1.MemberPrivilege != Org.MemberPrivilege {
			t.Error("Update Organization Failed")
		}
	}
}

//sysAdmin update Organization
func Test_sysAdminUpdateOrganization(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	Org.MemberPrivilege = dao.WRITE
	Org.Comment = "sysAdmin update"
	Org.URL = "url update"
	Org.Location = "land"
	statusCode, err := UpdateOrganizationTest(t, Org, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Failed.")
	}
}

//anyone (user2) update Organization
func Test_anyoneUpdateOrganization(t *testing.T) {

	Org.MemberPrivilege = dao.WRITE
	Org.Comment = "update Comment"
	Org.URL = "update url"
	Org.Location = "land"
	statusCode, err := UpdateOrganizationTest(t, Org, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update Organization Error.")
	}
}

//Non-Existed user update Organization
func Test_NonExistUserUpdateOrganization(t *testing.T) {

	user3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	Org.MemberPrivilege = dao.WRITE
	Org.Comment = "update Comment"
	Org.URL = "update url"
	Org.Location = "land"
	statusCode, err := UpdateOrganizationTest(t, Org, user3.Name, user3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update Organization Error.")
	}
}

//orgAdmin (user1) update Non Exist Organization
func Test_orgAdminUpdateNonExistOrganization(t *testing.T) {

	org1 := &dao.Organization{
		Name:            "NonExist",
		Email:           "admin@gmail.com",
		Comment:         "orgAdmin update",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, org1, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update Organization Error.")
	}
}

//Organization's Email,Comment,URL or Location is empty update test
func Test_UpdateWithEmptyFieldsOrganization(t *testing.T) {

	org := &dao.Organization{
		Name: "huawei",
		//Email:           "admin@gmail.com",
		//Comment:         "Comment orgjson,
		//URL:             "URL",
		//Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	//Email  empty
	//Comment empty
	org.URL = "url test"
	org.Location = "land"
	statusCode, err := UpdateOrganizationTest(t, org, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Error.")
	}

	//Email:  ""
	org.Email = " "
	org.Comment = "update org"
	statusCode, err = UpdateOrganizationTest(t, org, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Error.")
	}

	org.Email = "admin@gmail.com"
	stCode, err := UpdateOrganizationTest(t, org, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(stCode)
	if stCode != 200 {
		t.Fatal("Update Organization Failed.")
	}
}

// delete
func DeleteOrganizationTest(t *testing.T, OrgName, userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/organization/"+OrgName, nil)
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

//Anyone delete Organization
func Test_AnyoneDeleteOrganization(t *testing.T) {

	statusCode, err := DeleteOrganizationTest(t, Org.Name, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgadmin delete Non Existed Organization
func Test_orgAdminDeleteNonExistedOrganization(t *testing.T) {

	// Init organization struct
	org3 := dao.Organization{
		Name:            "huawei.cn",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}

	statusCode, err := DeleteOrganizationTest(t, org3.Name, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgMember delete Organization
func Test_orgMemberDeleteOrganization(t *testing.T) {

	//add user2 to Organization
	oum := &dao.OrganizationUserMap{
		User: User2,
		Role: dao.ORGMEMBER,
		Org:  Org,
	}
	if err := oum.Save(); err != nil {
		t.Error(err)
	}

	//orgMember(user2) delete Organization
	statusCode, err := DeleteOrganizationTest(t, Org.Name, User2.Name, User2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgadmin delete Organization
func Test_orgAdminDeleteOrganization(t *testing.T) {

	//delete Organization
	statusCode, err := DeleteOrganizationTest(t, Org.Name, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Delete Organization Failed.")
	}

	//query organization
	org1 := &dao.Organization{Name: Org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else {
		if exist {
			t.Error("Delete Organization Failed")
		}
	}
}

//clear test
func Test_organizationClear(t *testing.T) {

	if err := User1.Delete(); err != nil {
		t.Error(err)
	}
	if err := User2.Delete(); err != nil {
		t.Error(err)
	}
}
