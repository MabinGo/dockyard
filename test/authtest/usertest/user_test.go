package usertest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	user1, user2, sysAdmin, sysAdmin1, sysAdmin2 *dao.User
	org                                          *dao.Organization
)

func Test_UserInit(t *testing.T) {

	authtest.GetConfig()

	//1. create user1 user2
	user1 = &dao.User{
		Name:     "admin1",
		Email:    "admin1@gmail.com",
		Password: "admin123",
		//		RealName: "admin1",
		//		Comment:  "Comment",
		Status: 0,
	}
	authtest.SignUp(t, user1)

	user2 = &dao.User{
		Name:     "test1",
		Email:    "test1@gmail.com",
		Password: "test1",
		//		RealName: "test1",
		//		Comment:  "Comment",
		Status: 0,
	}
	authtest.SignUp(t, user2)

	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
		Role:     dao.SYSADMIN,
	}

	org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		MemberPrivilege: dao.WRITE,
	}
}

//
//1.============================== Test CreateUser API ==============================
//
//sysMember create sysAdmin1
func Test_SysMemberCreateUser(t *testing.T) {

	sysAdmin1 := &dao.User{
		Name:     "sys_admin1",
		Email:    "sysAdmin1@gmail.com",
		Password: "sysAdmin1",
		RealName: "sysAdmin1",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSADMIN,
	}

	statusCode, err := CreateUser(t, sysAdmin1, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create User Error.")
	}
}

//Super sysAdmin create sysAdmin1
func Test_SuperSysAdminCreateSysAdmin(t *testing.T) {

	sysAdmin1 = &dao.User{
		Name:     "sys_admin1",
		Email:    "sysAdmin1@gmail.com",
		Password: "sysAdmin1",
		RealName: "sysAdmin1",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSADMIN,
	}

	statusCode, err := CreateUser(t, sysAdmin1, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create User Failed.")
	}
}

// sysAdmin1 create sysAdmin2
func Test_SysAdminCreateSysAdmin(t *testing.T) {

	sysAdmin2 = &dao.User{
		Name:     "sys_admin2",
		Email:    "sysAdmin2@gmail.com",
		Password: "sysAdmin2",
		RealName: "sysAdmin2",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSADMIN,
	}

	statusCode, err := CreateUser(t, sysAdmin2, sysAdmin1.Name, sysAdmin1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Create User Error.")
	}
}

// Super sysAdmin  create Multi sysAdmin
func Test_SuperSysAdminCreateMultiSysAdmin(t *testing.T) {

	sysAdmin2 = &dao.User{
		Name:     "sys_admin2",
		Email:    "sysAdmin2@gmail.com",
		Password: "sysAdmin2",
		RealName: "sysAdmin2",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSADMIN,
	}

	statusCode, err := CreateUser(t, sysAdmin2, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Create User Error.")
	}
}

//
//2.============================== Test DeleteUser API ==============================
//
// sysMember delete sysMember
func Test_SysMemberDeleteSysMember(t *testing.T) {

	statusCode, err := authtest.DeleteUser(t, user2.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete User Error.")
	}
}

// sysMember delete sysAdmin1
func Test_SysMemberDeleteSysAdmin(t *testing.T) {

	statusCode, err := authtest.DeleteUser(t, sysAdmin1.Name, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Delete User Error.")
	}
}

//sysAdmin2 delete sysAdmin1
func Test_SysAdminDeleteSysAdmin(t *testing.T) {

	stCode, err := authtest.DeleteUser(t, sysAdmin1.Name, sysAdmin2.Name, sysAdmin2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(stCode)
	if stCode == 200 {
		t.Fatal("Delete User Error.")
	}
}

//Super sysAdmin delete sysAdmin1
func Test_SuperSysAdminDeleteSysAdmin(t *testing.T) {

	statusCode, err := authtest.DeleteUser(t, sysAdmin1.Name, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Delete User Error.")
	}
}

//
//3.============================== Test UpdateUser API ==============================
//
//sysAdmin update non-exist User
func Test_SysAdminUpdateNonExistUser(t *testing.T) {

	u := map[string]interface{}{
		"name":     "user3",
		"email":    "user3@gmail.com",
		"realname": "user3",
		"comment":  "Comment",
		"password": "123456",
	}

	statusCode, err := UpdateUser(t, u, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update User Error.")
	}
}

//sysAdmin update user1
func Test_SysAdminUpdateUser(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	u := map[string]interface{}{
		"name":     user1.Name,
		"email":    "user1@gmail.com",
		"realname": "user1",
		"comment":  "update map struct",
	}

	statusCode, err := UpdateUser(t, u, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update User Failed.")
	}
}

//user1 update other User
func Test_UserUpdateOtherUser(t *testing.T) {
	u := map[string]interface{}{
		"name":     user2.Name,
		"email":    "user2@gmail.com",
		"realname": "user2",
		"comment":  "update user2",
	}

	statusCode, err := UpdateUser(t, u, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update User Error.")
	}
}

//user2 update Password
func Test_UserUpdateUserPassword(t *testing.T) {

	oldPassword := user2.Password
	newPassword := "123456abcd"
	u := map[string]interface{}{
		"name":     user2.Name,
		"realname": "user2",
		"comment":  "update user2",
		"password": newPassword,
	}
	statusCode, err := UpdateUser(t, u, user2.Name, oldPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update User Failed.")
	}

	u["password"] = "test1"
	u["comment"] = "update password"
	statusCode1, err := UpdateUser(t, u, user2.Name, newPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode1)
	if statusCode1 != 200 {
		t.Fatal("Update User Failed.")
	}
}

//sysAdmin use wrong parameter update user1
func Test_UseWrongParameterUpdateUser(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	u := map[string]interface{}{
		"name":        user1.Name,
		"email":       "user1@gmail.com",
		"realname":    "user1",
		"comment":     "update map struct",
		"wrongstring": "wrong strings",
	}

	statusCode, err := UpdateUser(t, u, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update User Failed.")
	}
}

// Create User API
func CreateUser(t *testing.T, user *dao.User, userName, password string) (int, error) {

	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", setting.ListenMode+"://"+authtest.Domains+"/uam/user/add", bytes.NewBuffer(body))
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

// update User API
func UpdateUser(t *testing.T, user map[string]interface{}, userName, password string) (int, error) {

	body, _ := json.Marshal(user)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+authtest.Domains+"/uam/user/update", bytes.NewBuffer(body))
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

//
//4.============================== Test DeactiveUser API ==============================
//
func Test_DeactiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user2.Name
	if err := authtest.MethodFunc("DELETE", url, nil, user1); err == nil {
		t.Errorf("DeactiveUser error: only system admin can deactive user")
	}
	tearDownTest(t)
}

func Test_CreateOrgAfterDeactiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user2.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	//create org
	statusCode, err := authtest.CreateOrganization(t, org, user2.Name, user2.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("DeactiveUser Failed")
	}
	tearDownTest(t)
}

func Test_PushImageAfterDeactiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user2.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	// push images
	if err := authtest.PushPullImage(user2.Name, user2); err == nil {
		t.Errorf("DeactiveUser Failed")
	}
	tearDownTest(t)
}

//
//5.============================== Test ActiveUser API ==============================
//
func Test_ActiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user2.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/user/active/" + user2.Name
	if err := authtest.MethodFunc("PUT", url, nil, user1); err == nil {
		t.Errorf("ActiveUser error: Only system admin can active user")
	}
	tearDownTest(t)
}

func Test_SysAdminActiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/user/active/" + user1.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveUser error: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_CreateOrgAfterActiveUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/user/active/" + user1.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveUser error: %v", err.Error())
	}

	//create org
	statusCode, err := authtest.CreateOrganization(t, org, user1.Name, user1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateOrganization Failed")
	}
	tearDownTest(t)
}

func Test_RootAdminActiveSysAdminUser(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + sysAdmin2.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}
	if err := authtest.PushPullImage(sysAdmin2.Name, sysAdmin2); err == nil {
		t.Errorf("DeactiveUser: %s Error", sysAdmin2.Name)
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/user/active/" + sysAdmin2.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveUser: %v", err.Error())
	}
	if err := authtest.PushPullImage(sysAdmin2.Name, sysAdmin2); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

func Test_PullImagesAfterActiveUser(t *testing.T) {
	setUpTest(t)

	url := setting.ListenMode + "://" + authtest.Domains + "/uam/user/deactive/" + user1.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveUser: %v", err.Error())
	}

	if err := authtest.PushPullImage(user1.Name, user2); err != nil {
		if !strings.HasPrefix(err.Error(), "Push") {
			t.Error(err)
		}
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/user/active/" + user1.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveUser error: %v", err.Error())
	}
	tearDownTest(t)
}

func setUpTest(t *testing.T) {
	authtest.SignUp(t, user1)
	authtest.SignUp(t, user2)
}

func tearDownTest(t *testing.T) {
	authtest.DeleteUser(t, user1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user2.Name, sysAdmin.Name, sysAdmin.Password)
}

//clear test
func Test_UserClear(t *testing.T) {
	authtest.DeleteOrganization(t, org.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user1.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, user2.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, sysAdmin2.Name, sysAdmin.Name, sysAdmin.Password)
}
