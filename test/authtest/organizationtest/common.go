package organizationtest

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
)

var (
	sysAdmin, orgAdmin, orgMember, teamAdmin, user1, user2 *dao.User
	org, org1, orgHw                                       *dao.Organization
	oum1, oum2                                             *controller.OrganizationUserMapJSON
	repoEx                                                 *controller.RepositoryJSON
)

func organizationtestInit(t *testing.T) {
	authtest.GetConfig()
	authtest.OpenDB(t)

	sysAdmin = &dao.User{
		Name:     "root",
		Password: "root",
		RealName: "root",
	}
	orgAdmin = &dao.User{
		Name:     "org_admin",
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
		Name:     "team_admin",
		Email:    "team_admin@huawei.com",
		Password: "123456",
	}
	user1 = &dao.User{
		Name:     "admin",
		Email:    "admin@gmail.com",
		Password: "admin",
		RealName: "admin",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	user2 = &dao.User{
		Name:     "test",
		Email:    "test@gmail.com",
		Password: "test",
		RealName: "test",
		Comment:  "Comment",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	//Init organization struct
	org = &dao.Organization{
		Name:            "huawei",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
	}
	org1 = &dao.Organization{
		Name:            "huawei.cn",
		MemberPrivilege: dao.WRITE,
	}
	orgHw = &dao.Organization{
		Name:            "huawei.com",
		Email:           "admin@gmail.com",
		Comment:         "Comment",
		URL:             "URL",
		Location:        "Location",
		MemberPrivilege: dao.WRITE,
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

	repoEx = &controller.RepositoryJSON{
		Name:    "busybox",
		OrgName: org1.Name,
	}

	authtest.SignUp(t, orgAdmin)
	authtest.SignUp(t, orgMember)
	authtest.SignUp(t, teamAdmin)
	authtest.SignUp(t, user1)
	authtest.SignUp(t, user2)
}

func setUpTest(t *testing.T) {
	authtest.CreateOrganization(t, org1, orgAdmin.Name, orgAdmin.Password)
	authtest.CreateRepositoryEx(t, repoEx, orgAdmin)
	authtest.AddUserToOrganization(t, oum1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum2, orgAdmin.Name, orgAdmin.Password)
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
