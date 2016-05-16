package authtest

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	sysAdmin, orgAdmin, orgMember, teamAdmin, teamMember *dao.User
	org                                                  *dao.Organization
	team                                                 *controller.TeamJSON
)

func Test_RepositoryInit(t *testing.T) {
	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
		RealName: "root",
	}
	orgAdmin = &dao.User{
		Name:     "admin",
		Email:    "admin@huawei.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "commnet",
	}
	orgMember = &dao.User{
		Name:     "member",
		Email:    "member@huawei.com",
		Password: "123456",
	}
	teamAdmin = &dao.User{
		Name:     "teamadmin",
		Email:    "teamadmin@huawei.com",
		Password: "123456",
	}
	teamMember = &dao.User{
		Name:     "user",
		Email:    "user@huawei.com",
		Password: "123456",
	}
	org = &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	team = &controller.TeamJSON{
		TeamName: "team",
		OrgName:  "huawei",
	}
	oum1 := &controller.OrganizationUserMapJSON{
		UserName: orgMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	oum2 := &controller.OrganizationUserMapJSON{
		UserName: teamAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	oum3 := &controller.OrganizationUserMapJSON{
		UserName: teamMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	tum1 := &controller.TeamUserMapJSON{
		TeamName: "team",
		OrgName:  "huawei",
		UserName: teamAdmin.Name,
		Role:     dao.TEAMADMIN,
	}
	tum2 := &controller.TeamUserMapJSON{
		TeamName: "team",
		OrgName:  "huawei",
		UserName: teamMember.Name,
		Role:     dao.TEAMMEMBER,
	}
	signUp(orgAdmin, t)
	signUp(orgMember, t)
	signUp(teamAdmin, t)
	signUp(teamMember, t)
	createOrganization(org, t)
	addUserToOrganization(oum1, t)
	addUserToOrganization(oum2, t)
	addUserToOrganization(oum3, t)
	createTeam(team, t)
	addUserToTeam(tum1, t)
	addUserToTeam(tum2, t)
}

// Create repository test
func Test_UserCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgAdmin.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_SysAdminCreateMemberRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "testSys",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgMember.Name,
	}

	statusCode, err := createRepository(repo, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_UserCreateExistedRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgAdmin.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_UserCreateEmptyRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		UserName: orgAdmin.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_UserCreateMissingIsPublicRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test1",
		UserName: orgAdmin.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_NonExistedUserCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "none",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: "nonexisted",
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_UserAndOrgExistedCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "existed",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org.Name,
		UserName: orgAdmin.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_UserAndOrgNonExistedCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "nonexisted",
		IsPublic: true,
		Comment:  "this is a repo",
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_OrgCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "testOrg",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_SysAdminCreateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "testAdmin",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org.Name,
	}

	statusCode, err := createRepository(repo, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_OrgCreateMissingIspublicRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:    "test1",
		OrgName: org.Name,
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_NonExistedOrgCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "none",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  "nonexisted",
	}

	statusCode, err := createRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_NoOrgRightToCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test2",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org.Name,
	}

	statusCode, err := createRepository(repo, orgMember.Name, orgMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("CreateRepository Failed.")
	}
}

func Test_TeamAddRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "test1",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, teamAdmin.Name, teamAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("AddRepository Failed.")
	}
}

func Test_SysAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "testAdmin",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("AddRepository Failed.")
	}
}

func Test_OrgMemberAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "testOrg",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, orgMember.Name, orgMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("AddRepository Failed.")
	}
}

func Test_OrgAdminAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "testOrg",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("AddRepository Failed.")
	}
}

func Test_TeamAddNonexistRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "nonexist",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, teamAdmin.Name, teamAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("AddRepository Failed.")
	}
}

func Test_NoRightToAddRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team.OrgName,
		RepoName: "test",
		TeamName: team.TeamName,
		Permit:   dao.WRITE,
	}

	statusCode, err := addRepository(repo, teamMember.Name, teamMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("AddRepository Failed.")
	}
}

// Delete Test
func Test_NamespaceEmptyDeleteRepository(t *testing.T) {
	repository := "testOrg"
	statusCode, err := deleteRepository("", repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("DeleteRepository Failed.")
	}
}

func Test_NonRightUserDeleteRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "test"
	statusCode, err := deleteRepository(username, repository, orgMember.Name, orgMember.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("UserDeleteRepository Failed.")
	}
}

func Test_UserDeleteNonExistedRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "nonexisted"
	statusCode, err := deleteRepository(username, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("UserDeleteRepository Failed.")
	}
}

func Test_UserDeleteRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "test"
	statusCode, err := deleteRepository(username, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("UserDeleteRepository Failed.")
	}
}

func Test_SysAdminDeleteRepository(t *testing.T) {
	username := orgMember.Name
	repository := "testSys"
	statusCode, err := deleteRepository(username, repository, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("UserDeleteRepository Failed.")
	}
}

func Test_NonExistedOrgDeleteRepository(t *testing.T) {
	orgname := "nonexisted"
	repository := "testOrg"
	statusCode, err := deleteRepository(orgname, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("OrgDeleteRepository Failed.")
	}
}

func Test_NonRightOrgDeleteRepository(t *testing.T) {
	orgname := org.Name
	repository := "test"
	statusCode, err := deleteRepository(orgname, repository, orgMember.Name, orgMember.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("OrgDeleteRepository Failed.")
	}
}

func Test_OrgDeleteNonExistedRepository(t *testing.T) {
	orgname := org.Name
	repository := "nonexisted"
	statusCode, err := deleteRepository(orgname, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("OrgDeleteRepository Failed.")
	}
}

func Test_TeamRemoveNonexistRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "nonexist"

	statusCode, err := removeRepository(orgname, team, repository, teamAdmin.Name, teamAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_RemoveNonexistTeamRepository(t *testing.T) {
	orgname := team.OrgName
	team := "nonexisted"
	repository := "test1"

	statusCode, err := removeRepository(orgname, team, repository, teamAdmin.Name, teamAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_SysRemoveRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "testAdmin"

	statusCode, err := removeRepository(orgname, team, repository, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_OrgMemberRemoveRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "testOrg"

	statusCode, err := removeRepository(orgname, team, repository, orgMember.Name, orgMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_OrgAdminRemoveRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "testOrg"

	statusCode, err := removeRepository(orgname, team, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_TeamMemberRemoveRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "test1"

	statusCode, err := removeRepository(orgname, team, repository, teamMember.Name, teamMember.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode == 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_TeamAdminRemoveRepository(t *testing.T) {
	orgname := team.OrgName
	team := team.TeamName
	repository := "test1"

	statusCode, err := removeRepository(orgname, team, repository, teamAdmin.Name, teamAdmin.Password)
	if err != nil {
		t.Error(err)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("RemoveRepository Failed.")
	}
}

func Test_OrgDeleteRepository(t *testing.T) {
	orgname := org.Name
	repository := "testOrg"
	statusCode, err := deleteRepository(orgname, repository, orgAdmin.Name, orgAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("OrgDeleteRepository Failed.")
	}
}

func Test_SysAdminDeleteOrgRepository(t *testing.T) {
	orgname := org.Name
	repository := "testAdmin"
	statusCode, err := deleteRepository(orgname, repository, sysAdmin.Name, sysAdmin.Password)
	if err != nil {
		t.Error(t)
	}
	t.Log(statusCode)
	if statusCode != 200 {
		t.Fatal("UserDeleteRepository Failed.")
	}
}

func Test_RepositoryClear(t *testing.T) {
	deleteOrganization(org, t)
	deleteRepository(orgAdmin.Name, "test1", orgAdmin.Name, orgAdmin.Password)
}

func signUp(user *dao.User, t *testing.T) {
	b, err := json.Marshal(user)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/user/signup", strings.NewReader(string(b)))
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
	t.Log(resp)
}

func createOrganization(org *dao.Organization, t *testing.T) {
	b, err := json.Marshal(org)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/organization", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func deleteOrganization(org *dao.Organization, t *testing.T) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/organization/"+org.Name, nil)
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func addUserToOrganization(oumJSON *controller.OrganizationUserMapJSON, t *testing.T) {
	b, err := json.Marshal(oumJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/organization/adduser", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func createTeam(teamJson *controller.TeamJSON, t *testing.T) {
	b, err := json.Marshal(teamJson)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/team", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func deleteTeam(teamJson *controller.TeamJSON, t *testing.T) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/team/"+teamJson.OrgName+"/"+teamJson.TeamName, nil)
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func addUserToTeam(tum *controller.TeamUserMapJSON, t *testing.T) {
	b, err := json.Marshal(tum)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/team/adduser", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(orgAdmin.Name, orgAdmin.Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func createRepository(repo *controller.RepositoryJSON, userName, password string) (int, error) {
	b, err := json.Marshal(repo)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/repository/", strings.NewReader(string(b)))
	if err != nil {
		return -1, err
	}
	req.SetBasicAuth(userName, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func deleteRepository(namespace, repository, userName, password string) (int, error) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/repository/"+namespace+"/"+repository, nil)
	if err != nil {
		return -1, err
	}
	req.SetBasicAuth(userName, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func addRepository(repo *controller.TeamRepositoryMapJSON, userName, password string) (int, error) {
	b, err := json.Marshal(repo)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", setting.ListenMode+"://"+Domains+"/uam/team/addrepository", strings.NewReader(string(b)))
	if err != nil {
		return -1, err
	}
	req.SetBasicAuth(userName, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func removeRepository(organization, team, repository, userName, password string) (int, error) {
	req, err := http.NewRequest("DELETE", setting.ListenMode+"://"+Domains+"/uam/team/removerepository/"+organization+"/"+team+"/"+repository, nil)
	if err != nil {
		return -1, err
	}
	req.SetBasicAuth(userName, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}
