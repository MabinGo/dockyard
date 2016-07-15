package authtest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
	"github.com/containerops/dockyard/utils/setting"
)

var openDBFlag = false

//OpenDB connect DB
func OpenDB(t *testing.T) {

	if openDBFlag {
		return
	}
	if err := db.RegisterDriver(setting.DBDriver); err != nil {
		t.Error(err)
	} else {
		db.Drv.RegisterModel(new(dao.User), new(dao.Organization),
			new(dao.RepositoryEx), new(dao.OrganizationUserMap),
			new(dao.Team), new(dao.TeamUserMap), new(dao.TeamRepositoryMap))
		err := db.Drv.InitDB(setting.DBDriver, setting.DBUser, setting.DBPasswd, setting.DBURI, setting.DBName, 0)
		if err != nil {
			t.Error(err)
		}
	}
	openDBFlag = true
}

//SignUp Api
func SignUp(t *testing.T, user *dao.User) {

	b, err := json.Marshal(user)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/user/signup", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	if resp != nil {
		t.Log(resp.StatusCode)
	}
}

// DeleteUser Api
func DeleteUser(t *testing.T, delUser, userName, password string) (int, error) {

	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/user/"+delUser, nil)
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

//CreateOrganization Api
func CreateOrganization(t *testing.T, org *dao.Organization, username, password string) (int, error) {
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

//DeleteOrganization Api
func DeleteOrganization(t *testing.T, OrgName, userName, password string) (int, error) {

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

//AddUserToOrganization Api
func AddUserToOrganization(t *testing.T, oumJSON *controller.OrganizationUserMapJSON,
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

//CreateTeam Api
func CreateTeam(t *testing.T, teamJSON *controller.TeamJSON, username, password string) (int, error) {
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

//AddUserToTeam Api
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

//AddRepoToTeam Api
func AddRepoToTeam(t *testing.T, trm *controller.TeamRepositoryMapJSON, user *dao.User) {
	b, err := json.Marshal(trm)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/team/addrepository"
	if err = MethodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

//CreateRepositoryEx Api
func CreateRepositoryEx(t *testing.T, repo *controller.RepositoryJSON, user *dao.User) {
	b, err := json.Marshal(repo)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/repository"
	if err := MethodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

//PushPullImage namespace: user name or organization name
func PushPullImage(namespace string, user *dao.User) error {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + namespace + "/" + "repo" + ":" + "latest"
	cmd := exec.Command(DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
	if buf, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Docker login faild: [Error]%v %s", err, string(buf))
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		return fmt.Errorf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if buf, err := exec.Command(DockerBinary, "push", repoDest).CombinedOutput(); err != nil {
		return fmt.Errorf("Push %v failed:[Error]%v %s", repoDest, err, string(buf))
	}
	if buf, err := exec.Command(DockerBinary, "pull", repoDest).CombinedOutput(); err != nil {
		return fmt.Errorf("Pull %v failed:[Error]%v %s", repoDest, err, string(buf))
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		return fmt.Errorf("Docker logout failed:[Error]%v", err)
	}

	return nil
}

//DockyardRestart : restart app for test
func DockyardRestart(t *testing.T) error {
	cmd := exec.Command(DockyardPath + "/test/authtest/restart.sh")
	cmd.Dir = DockyardPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("start dockyard faild: [Error]%v", err)
	}
	time.Sleep(1 * time.Second)
	GetConfig()
	return nil
}

//MethodFunc common func
func MethodFunc(method string, url string, body io.Reader, user *dao.User) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(user.Name, user.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("HttpRespose :%d", resp.StatusCode)
	}

	return nil
}
