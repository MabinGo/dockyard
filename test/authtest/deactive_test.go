package authtest

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	sysAdmin, orgAdmin, orgManager, orgMember, teamAdmin, teamMember *dao.User
	org1, org2                                                       *dao.Organization
	team1, team2                                                     *controller.TeamJSON
	oum1, oum2, oum3, oum4                                           *controller.OrganizationUserMapJSON
	tum1, tum2, tum3                                                 *controller.TeamUserMapJSON
	repo_ex                                                          *controller.RepositoryJSON
	trm                                                              *controller.TeamRepositoryMapJSON
)

func Test_InactiveInit(t *testing.T) {
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
	orgManager = &dao.User{
		Name:     "manager",
		Email:    "manager@huawei.com",
		Password: "manager",
		RealName: "manager",
		Comment:  "commnet",
	}
	orgMember = &dao.User{
		Name:     "member",
		Email:    "member@huawei.com",
		Password: "123456",
	}
	teamAdmin = &dao.User{
		Name:     "team_admin",
		Email:    "team_admin@huawei.com",
		Password: "123456",
	}
	teamMember = &dao.User{
		Name:     "team_member",
		Email:    "team_member@huawei.com",
		Password: "123456",
	}
	org1 = &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.READ,
	}
	org2 = &dao.Organization{
		Name:            "360",
		MemberPrivilege: dao.WRITE,
	}
	repo_ex = &controller.RepositoryJSON{
		Name:    "busybox",
		OrgName: org1.Name,
	}
	team1 = &controller.TeamJSON{
		TeamName: "team1",
		OrgName:  "huawei",
	}
	team2 = &controller.TeamJSON{
		TeamName: "team2",
		OrgName:  "huawei",
	}
	oum1 = &controller.OrganizationUserMapJSON{
		UserName: orgMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org1.Name,
	}
	oum2 = &controller.OrganizationUserMapJSON{
		UserName: teamAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org1.Name,
	}
	oum3 = &controller.OrganizationUserMapJSON{
		UserName: teamMember.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org1.Name,
	}
	oum4 = &controller.OrganizationUserMapJSON{
		UserName: orgAdmin.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org2.Name,
	}
	tum1 = &controller.TeamUserMapJSON{
		TeamName: team1.TeamName,
		OrgName:  team1.OrgName,
		UserName: teamAdmin.Name,
		Role:     dao.TEAMADMIN,
	}
	tum2 = &controller.TeamUserMapJSON{
		TeamName: team1.TeamName,
		OrgName:  team1.OrgName,
		UserName: teamMember.Name,
		Role:     dao.TEAMMEMBER,
	}
	tum3 = &controller.TeamUserMapJSON{
		TeamName: team2.TeamName,
		OrgName:  team2.OrgName,
		UserName: teamMember.Name,
		Role:     dao.TEAMADMIN,
	}
	trm = &controller.TeamRepositoryMapJSON{
		OrgName:  org1.Name,
		RepoName: repo_ex.Name,
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}
	signUp(orgAdmin, t)
	signUp(orgMember, t)
	signUp(orgManager, t)
	signUp(teamAdmin, t)
	signUp(teamMember, t)
}

func Test_DeactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactive/" + org1.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrganization:", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactive/" + org1.Name
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrganization:", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactive/" + org1.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveOrganization:", err.Error())
	}

	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveOrganization")
	}
	tearDownTest(t)
}

func Test_DeactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap:", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveOrgUserMap:", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveOrgUserMap:", err.Error())
	}
	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveOrgUserMap Error")
	}
	tearDownTest(t)
}

func Test_DeactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/repository/deactive/" + repo_ex.OrgName + "/" + repo_ex.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/repository/deactive/" + repo_ex.OrgName + "/" + repo_ex.Name
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/repository/deactive/" + repo_ex.OrgName + "/" + repo_ex.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveRepo: %v", err.Error())
	}
	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveRepo Error")
	}
	tearDownTest(t)
}

func Test_DeactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeam: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeam(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeam: %v", err.Error())
	}
	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeam Error")
	}
	tearDownTest(t)
}

func Test_DeactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamUserMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeamUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiveuser/" + tum2.OrgName + "/" + tum2.TeamName + "/" + tum2.UserName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamUserMap: %v", err.Error())
	}
	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeamUserMap Error")
	}
	tearDownTest(t)
}

func Test_DeactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := methodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamRepoMap: %v", err.Error())
	}
	if err := methodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeamRepoMap Error")
	}
	tearDownTest(t)
}

func Test_PushInactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PushInactiveTeamRepoMap: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err == nil {
		t.Errorf("PushInactiveTeamRepoMap Error")
	}
	tearDownTest(t)
}

func Test_PullInactiveTeam(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + Domains + "/uam/team/deactive/" + team1.OrgName + "/" + team1.TeamName
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PullInactiveTeam: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err == nil {
		t.Errorf("PushInactiveTeamRepoMap Error")
	}
	tearDownTest(t)
}

func Test_PullInactiveRepo(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + Domains + "/uam/repository/deactive/" + repo_ex.OrgName + "/" + repo_ex.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PullInactiveRepo: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "Pull") {
			t.Error(err)
		}
	}
	tearDownTest(t)
}

func Test_PushInactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PushInactiveOrgUserMap: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "Push") {
			t.Error(err)
		}
	}
	tearDownTest(t)
}

func Test_PullInactiveOrgUserMap(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactiveuser/" + org1.Name + "/" + orgMember.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PullInactiveOrgUserMap: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "Pull") {
			t.Error(err)
		}
	}
	tearDownTest(t)
}

func Test_PushAfterInactiveOrganization(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactive/" + org1.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PushInactiveOrganization: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "Push") {
			t.Error(err)
		}
	}
	tearDownTest(t)
}

func Test_PullInactiveOrganization(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + Domains + "/uam/organization/deactive/" + org1.Name
	if err := methodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("PullInactiveOrganization: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "Pull") {
			t.Error(err)
		}
	}
	tearDownTest(t)
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
	t.Log(resp.StatusCode)
}

func setUpTest(t *testing.T) {
	createOrganization(org1, orgAdmin, t)
	createOrganization(org2, orgManager, t)
	createRepositoryEx(repo_ex, orgAdmin, t)
	addUserToOrganization(oum1, orgAdmin, t)
	addUserToOrganization(oum2, orgAdmin, t)
	addUserToOrganization(oum3, orgAdmin, t)
	addUserToOrganization(oum4, orgManager, t)
	createTeam(team1, orgAdmin, t)
	createTeam(team2, orgAdmin, t)
	addUserToTeam(tum1, orgAdmin, t)
	addUserToTeam(tum2, orgAdmin, t)
	addUserToTeam(tum3, orgAdmin, t)
	addRepoToTeam(trm, orgAdmin, t)
}

func tearDownTest(t *testing.T) {
	deleteOrganization(org1, orgAdmin, t)
	deleteOrganization(org2, orgManager, t)
}

func createOrganization(org *dao.Organization, user *dao.User, t *testing.T) {
	b, err := json.Marshal(org)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/organization"
	if err = methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func deleteOrganization(org *dao.Organization, user *dao.User, t *testing.T) {
	url := setting.ListenMode + "://" + Domains + "/uam/organization/" + org.Name
	if err := methodFunc("DELETE", url, nil, user); err != nil {
		t.Error(err)
	}
}

func createRepositoryEx(repo *controller.RepositoryJSON, user *dao.User, t *testing.T) {
	b, err := json.Marshal(repo)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/repository"
	if err := methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func deleteRepositoryEx(repo *controller.RepositoryJSON, user *dao.User, t *testing.T) {
	url := setting.ListenMode + "://" + Domains + "/uam/repository/" + repo.OrgName + "/" + repo.Name
	if err := methodFunc("DELETE", url, nil, user); err != nil {
		t.Error(err)
	}
}

func addUserToOrganization(oumJSON *controller.OrganizationUserMapJSON, user *dao.User, t *testing.T) {
	b, err := json.Marshal(oumJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/organization/adduser"
	if err := methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func createTeam(teamJson *controller.TeamJSON, user *dao.User, t *testing.T) {
	b, err := json.Marshal(teamJson)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/team"
	if err := methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func deleteTeam(teamJson *controller.TeamJSON, user *dao.User, t *testing.T) {
	url := setting.ListenMode + "://" + Domains + "/uam/team/" + teamJson.OrgName + "/" + teamJson.TeamName
	if err := methodFunc("DELETE", url, nil, user); err != nil {
		t.Error(err)
	}
}

func addUserToTeam(tum *controller.TeamUserMapJSON, user *dao.User, t *testing.T) {
	b, err := json.Marshal(tum)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/team/adduser"
	if err = methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func addRepoToTeam(trm *controller.TeamRepositoryMapJSON, user *dao.User, t *testing.T) {
	b, err := json.Marshal(trm)
	if err != nil {
		t.Error("marshal error.")
	}

	url := setting.ListenMode + "://" + Domains + "/uam/team/addrepository"
	if err = methodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		t.Error(err)
	}
}

func methodFunc(method string, url string, body io.Reader, user *dao.User) error {
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
		return fmt.Errorf("HttpRespose :", resp.StatusCode)
	}

	return nil
}

func pushImage(org *dao.Organization, user *dao.User) error {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + org.Name + "/" + repoBase
	cmd := exec.Command(DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		return fmt.Errorf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "push", repoDest).Run(); err != nil {
		return fmt.Errorf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		return fmt.Errorf("Docker logout failed:[Error]%v", err)
	}

	return nil
}

func pullImage(org *dao.Organization, user *dao.User) error {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + org.Name + "/" + repoBase
	cmd := exec.Command(DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		return fmt.Errorf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "pull", repoDest).Run(); err != nil {
		return fmt.Errorf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
		return fmt.Errorf("Docker logout failed:[Error]%v", err)
	}

	return nil
}
