package repositorytest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/test/authtest"
	"github.com/containerops/dockyard/utils/setting"
)

func Test_RepositoryInit(t *testing.T) {
	repositorytestInit(t)

	authtest.SignUp(t, orgAdmin)
	authtest.SignUp(t, orgMember)
	authtest.SignUp(t, teamAdmin)
	authtest.SignUp(t, teamMember)
	authtest.CreateOrganization(t, org1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum2, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToOrganization(t, oum3, orgAdmin.Name, orgAdmin.Password)
	authtest.CreateTeam(t, team1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToTeam(t, tum1, orgAdmin.Name, orgAdmin.Password)
	authtest.AddUserToTeam(t, tum2, orgAdmin.Name, orgAdmin.Password)
}

//
//1.============================== Test CreateRepository API ==============================
//
// Create repository test
func Test_UserCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgAdmin.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_SysAdminCreateMemberRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_sys",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgMember.Name,
	}

	if err := createRepository(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

func Test_UserCreateExistedRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgAdmin.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserCreateEmptyRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		UserName: orgAdmin.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserCreateMissingIsPublicRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test1",
		UserName: orgAdmin.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_NonExistedUserCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "none",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: "nonexisted",
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserAndOrgExistedCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "existed",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org1.Name,
		UserName: orgAdmin.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserAndOrgNonExistedCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "nonexisted",
		IsPublic: true,
		Comment:  "this is a repo",
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_OrgCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org1.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_SysAdminCreateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_admin",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org1.Name,
	}

	if err := createRepository(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

func Test_OrgCreateMissingIspublicRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:    "test1",
		OrgName: org1.Name,
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_NonExistedOrgCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "none",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  "nonexisted",
	}

	if err := createRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_NoOrgRightToCreateRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test2",
		IsPublic: true,
		Comment:  "this is a repo",
		OrgName:  org1.Name,
	}

	if err := createRepository(repo, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

//
//2.============================== Test team addRepository API ==============================
//
func Test_TeamAddRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, teamAdmin); err != nil {
		t.Error(err)
	}
}

func Test_SysAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test_admin",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

func Test_OrgMemberAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test_org",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_OrgAdminAddTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test_org",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_TeamAddNonexistRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "nonexist",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_NoRightToAddRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := addRepository(repo, teamMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

//
//3.============================== Test  Update Repository  API ==============================
//

// Update Repository Test
// orgAdmin update Org repository
func Test_orgAdminUpdateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: false,
		Comment:  "orgAdmin update repo",
		OrgName:  org1.Name,
	}

	if err := updateRepository(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

// sysAdmin update Org repository
func Test_sysAdminUpdateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: true,
		Comment:  "sysAdmin update repo",
		OrgName:  org1.Name,
	}

	if err := updateRepository(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

// orgMember update Org repository
func Test_orgMemberUpdateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: false,
		Comment:  "orgMember update repo",
		OrgName:  org1.Name,
	}

	if err := updateRepository(repo, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// non-exist user update Org repository
func Test_nonExistUserUpdateOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: false,
		Comment:  "orgMember update repo",
		OrgName:  org1.Name,
	}

	user := &dao.User{
		Name:     "tempuser",
		Password: "123456abc",
	}

	if err := updateRepository(repo, user); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// sysAdmin update Org non-exist repository
func Test_UpdateOrgNonExistRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "nonexistrepo",
		IsPublic: true,
		Comment:  "sysAdmin update repo",
		OrgName:  org1.Name,
	}

	if err := updateRepository(repo, sysAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

//  update non-exist Org repository
func Test_UpdateNonExistOrgRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "test_org",
		IsPublic: true,
		Comment:  "orgAdmin update repo",
		OrgName:  "nonexistorg",
	}

	if err := updateRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// update user repository
func Test_userUpdateRepository(t *testing.T) {
	//1.create user repo
	repo := &controller.RepositoryJSON{
		Name:     "user_repo",
		IsPublic: true,
		Comment:  "this is a repo",
		UserName: orgMember.Name,
	}
	createRepository(repo, orgMember)

	//2.update
	repo.IsPublic = true
	repo.Comment = "orgMember update repo"

	if err := updateRepository(repo, orgMember); err != nil {
		t.Error(err)
	}
}

// someone update other user repository
func Test_userUpdateOtherUserRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "user_repo",
		IsPublic: false,
		Comment:  "orgAdmin update repo",
		UserName: orgMember.Name,
	}

	if err := updateRepository(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// sysAdmin update user repository
func Test_sysAdminUpdateUserRepository(t *testing.T) {
	repo := &controller.RepositoryJSON{
		Name:     "user_repo",
		IsPublic: false,
		Comment:  "sysAdmin update repo",
		UserName: orgMember.Name, //UserName: sysAdmin.Name,
	}

	if err := updateRepository(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

//
//4.============================== Test  Update TeamRepositoryMap  API ==============================
//
//Update TeamRepositoryMap Test
// teamAdmin update team repository map
func Test_teamAdminUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.READ,
	}

	//update
	if err := updateTeamRepositoryMap(repo, teamAdmin); err != nil {
		t.Error(err)
	}
}

// orgAdmin update team repository map
func Test_orgAdminUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, orgAdmin); err != nil {
		t.Error(err)
	}
}

// sysAdmin update team repository map
func Test_sysAdminUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.READ,
	}

	if err := updateTeamRepositoryMap(repo, sysAdmin); err != nil {
		t.Error(err)
	}
}

// teamMember update team repository map
func Test_teamMemberUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, teamMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// other user update team repository map
func Test_orgMemberUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// orgAdmin update not right org team repository map
func Test_orgAdminUpdateNotRightTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  "emptyorg",
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// orgAdmin update non exist team repository map
func Test_orgAdminUpdateNonExistTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "nonexist",
		TeamName: team1.TeamName,
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// orgAdmin update non exist team repository map
func Test_orgAdminUpdateEmptyTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: "emptyteam",
		Permit:   dao.WRITE,
	}

	if err := updateTeamRepositoryMap(repo, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

// sysAdmin Use Wrong Parameter update team repository map
func Test_UseWrongParameterUpdateTeamRepository(t *testing.T) {
	repo := &controller.TeamRepositoryMapJSON{
		OrgName:  team1.OrgName,
		RepoName: "test1",
		TeamName: team1.TeamName,
		Permit:   dao.READ,
		Status:   1, // Status can not update
	}

	if err := updateTeamRepositoryMap(repo, sysAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

//
//5.============================== Test  Delete Repository  API ==============================
//

// Delete Test
func Test_NamespaceEmptyDeleteRepository(t *testing.T) {
	repository := "test_org"
	if err := deleteRepository("", repository, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_NonRightUserDeleteRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "test"
	if err := deleteRepository(username, repository, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserDeleteNonExistedRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "nonexisted"
	if err := deleteRepository(username, repository, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_UserDeleteRepository(t *testing.T) {
	username := orgAdmin.Name
	repository := "test"
	if err := deleteRepository(username, repository, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_SysAdminDeleteRepository(t *testing.T) {
	username := orgMember.Name
	repository := "test_sys"
	if err := deleteRepository(username, repository, sysAdmin); err != nil {
		t.Error(err)
	}
}

func Test_NonExistedOrgDeleteRepository(t *testing.T) {
	orgname := "nonexisted"
	repository := "test_org"
	if err := deleteRepository(orgname, repository, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_NonRightOrgDeleteRepository(t *testing.T) {
	orgname := org1.Name
	repository := "test"
	if err := deleteRepository(orgname, repository, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_OrgDeleteNonExistedRepository(t *testing.T) {
	orgname := org1.Name
	repository := "nonexisted"
	if err := deleteRepository(orgname, repository, orgAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_TeamRemoveNonexistRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "nonexist"

	if err := removeRepository(orgname, team, repository, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_RemoveNonexistTeamRepository(t *testing.T) {
	orgname := team1.OrgName
	team := "nonexisted"
	repository := "test1"

	if err := removeRepository(orgname, team, repository, teamAdmin); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_SysRemoveRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "test_admin"

	if err := removeRepository(orgname, team, repository, sysAdmin); err != nil {
		t.Error(err)
	}
}

func Test_OrgMemberRemoveRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "test_org"

	if err := removeRepository(orgname, team, repository, orgMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_OrgAdminRemoveRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "test_org"

	if err := removeRepository(orgname, team, repository, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_TeamMemberRemoveRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "test1"

	if err := removeRepository(orgname, team, repository, teamMember); err != nil {
		if !strings.HasPrefix(err.Error(), "HttpRespose") {
			t.Error(err)
		}
	}
}

func Test_TeamAdminRemoveRepository(t *testing.T) {
	orgname := team1.OrgName
	team := team1.TeamName
	repository := "test1"

	if err := removeRepository(orgname, team, repository, teamAdmin); err != nil {
		t.Error(err)
	}
}

func Test_OrgDeleteRepository(t *testing.T) {
	orgname := org1.Name
	repository := "test_org"
	if err := deleteRepository(orgname, repository, orgAdmin); err != nil {
		t.Error(err)
	}
}

func Test_SysAdminDeleteOrgRepository(t *testing.T) {
	orgname := org1.Name
	repository := "test_admin"
	if err := deleteRepository(orgname, repository, sysAdmin); err != nil {
		t.Error(err)
	}
}

//
//6.============================== Test DeactiveRepo API ==============================
//
func Test_DeactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveRepo: %v", err.Error())
	}
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveRepo Error")
	}
	tearDownTest(t)
}

//
//7.============================== Test ActiveRepo API ==============================
//
func Test_ActiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/repository/active/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/repository/active/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveRepo: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveRepo(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveRepo: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/repository/active/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveRepo: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveRepo Error")
	}
	tearDownTest(t)
}

func Test_PullActiveRepo(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/deactive/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveRepo: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/repository/active/" + repoEx.OrgName + "/" + repoEx.Name
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveRepo: %v", err.Error())
	}
	if err := pullImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	tearDownTest(t)
}

//================================================================================================

func createRepository(repo *controller.RepositoryJSON, user *dao.User) error {
	b, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/"
	if err = authtest.MethodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		return err
	}

	return nil
}

func updateRepository(repo *controller.RepositoryJSON, user *dao.User) error {
	b, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/update"
	if err := authtest.MethodFunc("PUT", url, strings.NewReader(string(b)), user); err != nil {
		return err
	}

	return nil
}

func updateTeamRepositoryMap(repo *controller.TeamRepositoryMapJSON, user *dao.User) error {
	b, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/updateteamrepositorymap"
	if err := authtest.MethodFunc("PUT", url, strings.NewReader(string(b)), user); err != nil {
		return err
	}

	return nil
}

func deleteRepository(namespace, repository string, user *dao.User) error {
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/repository/" + namespace + "/" + repository
	if err := authtest.MethodFunc("DELETE", url, nil, user); err != nil {
		return err
	}

	return nil
}

func addRepository(repo *controller.TeamRepositoryMapJSON, user *dao.User) error {
	b, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/addrepository"
	if err = authtest.MethodFunc("POST", url, strings.NewReader(string(b)), user); err != nil {
		return err
	}

	return nil
}

func removeRepository(organization, team, repository string, user *dao.User) error {
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/removerepository/" + organization + "/" + team + "/" + repository
	if err := authtest.MethodFunc("DELETE", url, nil, user); err != nil {
		return err
	}

	return nil
}

func Test_RepositoryClear(t *testing.T) {
	authtest.DeleteOrganization(t, org1.Name, orgAdmin.Name, orgAdmin.Password)
	deleteRepository(orgAdmin.Name, "test1", orgAdmin)
	deleteRepository(orgMember.Name, "user_repo", sysAdmin)
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamMember.Name, sysAdmin.Name, sysAdmin.Password)
}
