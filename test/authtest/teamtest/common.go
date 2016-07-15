package teamtest

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
)

var (
	sysAdmin, orgAdmin, orgMember, teamAdmin, teamMember *dao.User
	org1                                                 *dao.Organization
	team1, team2                                         *controller.TeamJSON
	oum1, oum2, oum3                                     *controller.OrganizationUserMapJSON
	tum1, tum2, tum3                                     *controller.TeamUserMapJSON
	repoEx                                               *controller.RepositoryJSON
	trm                                                  *controller.TeamRepositoryMapJSON
)

func teamtestInit(t *testing.T) {
	authtest.GetConfig()
	authtest.OpenDB(t)

	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
		RealName: "root",
	}
	orgAdmin = &dao.User{
		Name:     "adminorg",
		Email:    "adminorg@huawei.com",
		Password: "adminorg",
		RealName: "adminorg",
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
		Name:            "google",
		MemberPrivilege: dao.READ,
	}
	repoEx = &controller.RepositoryJSON{
		Name:    "busybox",
		OrgName: org1.Name,
	}
	team1 = &controller.TeamJSON{
		TeamName: "team1",
		OrgName:  org1.Name,
	}
	team2 = &controller.TeamJSON{
		TeamName: "team2",
		OrgName:  org1.Name,
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
		RepoName: repoEx.Name,
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	authtest.SignUp(t, orgAdmin)
	authtest.SignUp(t, orgMember)
	authtest.SignUp(t, teamAdmin)
	authtest.SignUp(t, teamMember)
}

func setUpTest(t *testing.T) {
	authtest.CreateOrganization(t, org1, orgAdmin.Name, orgAdmin.Password)
	authtest.CreateRepositoryEx(t, repoEx, orgAdmin)
	authtest.AddUserToOrganization(t, oum1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum2, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum3, orgAdmin.Name, orgAdmin.Password)
	authtest.CreateTeam(t, team1, orgAdmin.Name, orgAdmin.Password)
	authtest.CreateTeam(t, team2, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToTeam(t, tum1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToTeam(t, tum2, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToTeam(t, tum3, orgAdmin.Name, orgAdmin.Password)
	authtest.AddRepoToTeam(t, trm, orgAdmin)
}

func tearDownTest(t *testing.T) {
	authtest.DeleteOrganization(t, org1.Name, orgAdmin.Name, orgAdmin.Password)
}

func pushImage(org *dao.Organization, user *dao.User) error {
	repoBase := "busybox:latest"
	repoDest := authtest.Domains + "/" + org.Name + "/" + repoBase
	cmd := exec.Command(authtest.DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, authtest.Domains)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(authtest.DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		return fmt.Errorf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(authtest.DockerBinary, "push", repoDest).Run(); err != nil {
		return fmt.Errorf("Push %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(authtest.DockerBinary, "logout", authtest.Domains).Run(); err != nil {
		return fmt.Errorf("Docker logout failed:[Error]%v", err)
	}

	return nil
}

func pullImage(org *dao.Organization, user *dao.User) error {
	repoBase := "busybox:latest"
	repoDest := authtest.Domains + "/" + org.Name + "/" + repoBase
	cmd := exec.Command(authtest.DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, authtest.Domains)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker login faild: [Error]%v", err)
	}
	if err := exec.Command(authtest.DockerBinary, "tag", "-f", repoBase, repoDest).Run(); err != nil {
		return fmt.Errorf("Tag %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(authtest.DockerBinary, "pull", repoDest).Run(); err != nil {
		return fmt.Errorf("Pull %v failed:[Error]%v", repoDest, err)
	}
	if err := exec.Command(authtest.DockerBinary, "logout", authtest.Domains).Run(); err != nil {
		return fmt.Errorf("Docker logout failed:[Error]%v", err)
	}

	return nil
}
