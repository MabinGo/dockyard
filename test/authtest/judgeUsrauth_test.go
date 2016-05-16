package authtest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func TestUsrAuth(t *testing.T) {
	//First case: User is admin
	//This environment can not be constructed now
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Second case: User in organization map is admin auth
	usr := dao.User{
		Name:     "usr",
		Email:    "usr@mail.com",
		Password: "usr",
		//RealName string `orm:"size(100);null"`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   2,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "user" + "/" + "signup"
	rst, err := json.Marshal(usr)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	req, err := http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	org := dao.Organization{
		Name:            "org",
		MemberPrivilege: dao.WRITE,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/"
	rst, err = json.Marshal(org)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase := "busybox:latest"
	repoDest := Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd := exec.Command(DockerBinary, "login", "-u", usr.Name, "-p", usr.Password, "-e", usr.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Third case: User in organization map is member and write auth
	usrorgmem := dao.User{
		Name:     "usrorgmem",
		Email:    "usrorgmem@mail.com",
		Password: "usrorgmem",
		//RealName string `orm:"size(100);null"`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   2,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "user" + "/" + "signup"
	rst, err = json.Marshal(usrorgmem)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + org.Name
	rst, err = json.Marshal(org)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("DELETE", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	} else if rsp.StatusCode != 200 {
		t.Fatal("Http request error")
	}

	org = dao.Organization{
		Name:            "org",
		MemberPrivilege: dao.WRITE,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/"
	rst, err = json.Marshal(org)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	oumJSON := controller.OrganizationUserMapJSON{
		UserName: usrorgmem.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  "org",
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + "adduser" + "/"
	rst, err = json.Marshal(oumJSON)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repo := controller.RepositoryJSON{
		Name:     "repo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  org.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err = json.Marshal(repo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usrorgmem.Name, "-p", usrorgmem.Password, "-e", usrorgmem.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Fourth case: User in organization map is member and read auth
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + org.Name
	rst, err = json.Marshal(org)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("DELETE", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	} else if rsp.StatusCode != 200 {
		t.Fatal("Http request error")
	}

	org = dao.Organization{
		Name:            "org",
		MemberPrivilege: dao.READ,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/"
	rst, err = json.Marshal(org)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repo = controller.RepositoryJSON{
		Name:     "repo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  org.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err = json.Marshal(repo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usrorgmem.Name, "-p", usrorgmem.Password, "-e", usrorgmem.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Fifth case: Repository is exist and public, organization is exist, but usr is not in the organization
	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usrorgmem.Name, "-p", usrorgmem.Password, "-e", usrorgmem.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Sixth case: Organization is not exist, users operate their own namespace
	repoBase = "busybox:latest"
	repoDest = Domains + "/" + usr.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usr.Name, "-p", usr.Password, "-e", usr.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//Seventh case: Organization is not exist, users operate namespace what is not their own
	otherusr := dao.User{
		Name:     "otherusr",
		Email:    "otherusr@mail.com",
		Password: "otherusr",
		//RealName string `orm:"size(100);null"`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   2,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "user" + "/" + "signup"
	rst, err = json.Marshal(otherusr)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repo = controller.RepositoryJSON{
		Name:     "repo",
		IsPublic: true,
		Comment:  "comment",
		UserName: otherusr.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err = json.Marshal(repo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(otherusr.Name, otherusr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + otherusr.Name + "/" + repo.Name + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", otherusr.Name, "-p", otherusr.Password, "-e", otherusr.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	cmd = exec.Command(DockerBinary, "login", "-u", usr.Name, "-p", usr.Password, "-e", usr.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//eighth case: User in team map is admin auth
	repo = controller.RepositoryJSON{
		Name:     "repo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  org.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err = json.Marshal(repo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	team := controller.TeamJSON{
		TeamName: "team",
		Comment:  "comment",
		OrgName:  org.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/"
	rst, err = json.Marshal(team)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usr.Name, "-p", usr.Password, "-e", usr.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//ninth case: User in team map is member and write auth
	oumJSON = controller.OrganizationUserMapJSON{
		UserName: usrorgmem.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  "org",
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + "adduser" + "/"
	rst, err = json.Marshal(oumJSON)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	trm := controller.TeamRepositoryMapJSON{
		OrgName:  org.Name,
		RepoName: repo.Name,
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + "addrepository" + "/"
	rst, err = json.Marshal(trm)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	tum := controller.TeamUserMapJSON{
		TeamName: team.TeamName,
		OrgName:  org.Name,
		UserName: usrorgmem.Name,
		Role:     dao.TEAMMEMBER,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + "adduser" + "/"
	rst, err = json.Marshal(tum)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	} else if rsp.StatusCode != 200 {
		t.Fatal("Http request error")
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usrorgmem.Name, "-p", usrorgmem.Password, "-e", usrorgmem.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		t.Fatalf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}

	//tenth case: User in team map is member and read auth
	requestUrl = "http://" + Domains + "/" + "uam" + "/" + "team" + "/" + "removerepository" + "/" + org.Name + "/" + team.TeamName + "/" + repo.Name
	req, err = http.NewRequest("DELETE", requestUrl, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	trm = controller.TeamRepositoryMapJSON{
		OrgName:  org.Name,
		RepoName: repo.Name,
		TeamName: team.TeamName,
		Permit:   dao.READ,
	}
	requestUrl = "http://" + Domains + "/" + "uam" + "/" + "team" + "/" + "addrepository" + "/"
	rst, err = json.Marshal(trm)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	req, err = http.NewRequest("POST", requestUrl, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.SetBasicAuth(usr.Name, usr.Password)
	client = &http.Client{}
	rsp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}

	repoBase = "busybox:latest"
	repoDest = Domains + "/" + org.Name + "/" + "repo" + ":" + "latest"
	cmd = exec.Command(DockerBinary, "login", "-u", usrorgmem.Name, "-p", usrorgmem.Password, "-e", usrorgmem.Email, Domains)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		t.Fatalf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		t.Fatalf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		t.Fatalf("Docker logout failed:[Error]%v", err)
	}
}
