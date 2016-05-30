package authtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func Test_userInit(t *testing.T) {

	//1. create user1 user2
	User1 = &dao.User{
		Name:     "admin1",
		Email:    "admin1@gmail.com",
		Password: "admin123",
		//		RealName: "admin1",
		//		Comment:  "Comment",
		Status: 0,
	}
	signUp(User1, t)

	User2 = &dao.User{
		Name:     "test1",
		Email:    "test1@gmail.com",
		Password: "test1",
		//		RealName: "test1",
		//		Comment:  "Comment",
		Status: 0,
	}
	signUp(User2, t)
}

// update
func UpdateUserTest(t *testing.T, user *dao.User, username, password string) (int, error) {
	body, _ := json.Marshal(user)
	req, err := http.NewRequest("PUT", setting.ListenMode+"://"+Domains+"/uam/user/update", bytes.NewBuffer(body))
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

//sysAdmin update non-exist User
func Test_sysAdminUpdateNonExistUser(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	User3 := &dao.User{
		Name:     "user3",
		Email:    "user3@gmail.com",
		Password: "user3",
		RealName: "user3",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	User3.Email = "user@gmail.com"
	//User3.Password = "admin"
	User3.RealName = "user"
	User3.Comment = "update user3"
	statusCode, err := UpdateUserTest(t, User3, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update User Error.")
	}
}

//sysAdmin update User1
func Test_sysAdminUpdateUser(t *testing.T) {

	sysAdmin := &dao.User{
		Name:     "root",
		Password: "root",
	}

	User1.Email = "user1@gmail.com"
	User1.RealName = "user1"
	User1.Comment = "update user1"
	statusCode, err := UpdateUserTest(t, User1, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update User Failed.")
	}
}

//User1 update other User
func Test_userUpdateOtherUser(t *testing.T) {

	User2.Email = "user2@gmail.com"
	User2.RealName = "user2"
	User2.Comment = "update user2"
	statusCode, err := UpdateUserTest(t, User2, User1.Name, User1.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("Update User Error.")
	}
}

//User2 update Password
func Test_userUpdateUserPassword(t *testing.T) {

	oldPassword := User2.Password
	User2.Password = "user2"
	User2.RealName = "user2"
	User2.Comment = "update user2"
	statusCode, err := UpdateUserTest(t, User2, User2.Name, oldPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("Update User Failed.")
	}

	newPassword := User2.Password
	User2.Password = "test1"
	User2.Comment = "update password"
	statusCode1, err := UpdateUserTest(t, User2, User2.Name, newPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode1)
	if statusCode1 != 200 {
		t.Fatal("Update User Failed.")
	}

}

//clear test
func Test_userClear(t *testing.T) {

	if err := User1.Delete(); err != nil {
		t.Error(err)
	}
	if err := User2.Delete(); err != nil {
		t.Error(err)
	}
}
