package organizationtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
	"github.com/containerops/dockyard/utils/setting"
)

func Test_OrganizationInit(t *testing.T) {
	organizationtestInit(t)
}

//
//1.============================== Test CreateOrganization API ==============================
//
//user1 Create Organization
func Test_CreateOrganization(t *testing.T) {

	statusCode, err := authtest.CreateOrganization(t, org, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create Organization Failed.")
	}

	// query organization
	org1 := &dao.Organization{Name: org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if org1.Name != org.Name || org1.Email != org.Email ||
			org1.Comment != org.Comment || org1.URL != org.URL ||
			org1.Location != org.Location ||
			org1.MemberPrivilege != org.MemberPrivilege {
			t.Error("org's save is not same with get")
		}
	}
}

//user2 create same Organization with user1
func Test_ReCreateOrganization(t *testing.T) {

	statusCode, err := authtest.CreateOrganization(t, org, user2.Name, user2.Password)
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

	statusCode, err := authtest.CreateOrganization(t, org2, user3.Name, user3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create Organization Error.")
	}

}

//
//2.============================== Test UpdateOrganization API ==============================
//
// Update
func UpdateOrganizationTest(t *testing.T, org map[string]interface{}, username, password string) (int, error) {
	body, _ := json.Marshal(org)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+authtest.Domains+"/uam/organization/update", bytes.NewBuffer(body))
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
func Test_OrgAdminUpdateOrganization(t *testing.T) {

	orgm := map[string]interface{}{
		"name":            org.Name,
		"comment":         "orgAdmin update",
		"url":             "url test",
		"location":        "location",
		"memberprivilege": dao.READ,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Failed.")
	}

	// query organization
	org1 := &dao.Organization{Name: org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("org is not exitst")
	} else {
		if org1.Name != orgm["name"] || org1.Location != orgm["location"] ||
			org1.Comment != orgm["comment"] || org1.URL != orgm["url"] ||
			org1.MemberPrivilege != orgm["memberprivilege"] {
			t.Error("Update Organization Failed")
		}
	}
}

//sysAdmin update Organization
func Test_SysAdminUpdateOrganization(t *testing.T) {

	orgm := map[string]interface{}{
		"name":            org.Name,
		"comment":         "sysAdmin update",
		"url":             "url update",
		"location":        "land",
		"memberprivilege": dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Failed.")
	}
}

//anyone (user2) update Organization
func Test_AnyoneUpdateOrganization(t *testing.T) {

	orgm := map[string]interface{}{
		"name":            org.Name,
		"comment":         "update Comment",
		"url":             "update url",
		"location":        "land",
		"memberprivilege": dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, user2.Name, user2.Password)
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

	orgm := map[string]interface{}{
		"name":            org.Name,
		"comment":         "update user3",
		"url":             "update url",
		"location":        "land",
		"memberprivilege": dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, user3.Name, user3.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update Organization Error.")
	}
}

//orgAdmin (user1) update Non Exist Organization
func Test_OrgAdminUpdateNonExistOrganization(t *testing.T) {

	orgm := map[string]interface{}{
		"name":            "non_exist",
		"email":           "admin@gmail.com",
		"comment":         "orgAdmin update",
		"url":             "URL",
		"location":        "Location",
		"memberprivilege": dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, user1.Name, user1.Password)
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

	orgm := map[string]interface{}{
		"name": org.Name,
		//"email":           "admin@gmail.com",
		//"comment":         "orgAdmin update",
		"url":             "url test",
		"location":        "land",
		"memberprivilege": dao.WRITE,
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Error.")
	}

	//Email:  ""
	orgm["email"] = " "
	orgm["comment"] = "update org"
	statusCode, err = UpdateOrganizationTest(t, orgm, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update Organization Error.")
	}

	orgm["email"] = "admin@gmail.com"
	stCode, err := UpdateOrganizationTest(t, orgm, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(stCode)
	if stCode != 200 {
		t.Fatal("Update Organization Failed.")
	}
}

//sysAdmin Use wrong Parameter update Organization
func Test_UseWrongParameterUpdateOrganization(t *testing.T) {

	orgm := map[string]interface{}{
		"name":            org.Name,
		"comment":         "sysAdmin update",
		"memberprivilege": dao.WRITE,
		"wrongsrting":     "wrong srtings",
	}

	statusCode, err := UpdateOrganizationTest(t, orgm, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update Organization Error.")
	}
}

//
//3.============================== Test DeleteOrganization API ==============================
//
//Anyone delete Organization
func Test_AnyoneDeleteOrganization(t *testing.T) {

	statusCode, err := authtest.DeleteOrganization(t, org.Name, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgadmin delete Non Existed Organization
func Test_OrgAdminDeleteNonExistedOrganization(t *testing.T) {

	// Init organization struct
	org3 := dao.Organization{
		Name:            "huawei.cn",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}

	statusCode, err := authtest.DeleteOrganization(t, org3.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgMember delete Organization
func Test_OrgMemberDeleteOrganization(t *testing.T) {

	//add user2 to Organization
	oum := &dao.OrganizationUserMap{
		User: user2,
		Role: dao.ORGMEMBER,
		Org:  org,
	}
	if err := oum.Save(); err != nil {
		t.Error(err)
	}

	//orgMember(user2) delete Organization
	statusCode, err := authtest.DeleteOrganization(t, org.Name, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete Organization Error.")
	}
}

//orgadmin delete Organization
func Test_OrgAdminDeleteOrganization(t *testing.T) {

	//delete Organization
	statusCode, err := authtest.DeleteOrganization(t, org.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Delete Organization Failed.")
	}

	//query organization
	org1 := &dao.Organization{Name: org.Name}
	if exist, err := org1.Get(); err != nil {
		t.Error(err)
	} else {
		if exist {
			t.Error("Delete Organization Failed")
		}
	}
}

//
//4.============================== Test DeactiveOrganization API ==============================
//
func Test_DeactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveOrganization: %v", err.Error())
	}

	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveOrganization")
	}
	tearDownTest(t)
}

//
//5.============================== Test ActiveOrganization API ==============================
//
func Test_ActiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/active/" + org1.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrganization: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/active/" + org1.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveOrganization: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/active/" + org1.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrganization: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveOrganization Error")
	}
	tearDownTest(t)
}

func Test_PushAfterActiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PushInactiveOrganization: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/active/" + org1.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrganization: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

func Test_PullActiveOrganization(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/organization/deactive/" + org1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrganization: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/organization/active/" + org1.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveOrganization: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

//clear test
func Test_OrganizationClear(t *testing.T) {
	authtest.DeleteUser(t, user1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user2.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
}
