package teamtest

import (
	"testing"

	"github.com/containerops/dockyard/test/authtest"
	"github.com/containerops/dockyard/utils/setting"
)

func Test_TeamRepostitoryMapInit(t *testing.T) {
	teamtestInit(t)
}

//
//1.============================== Test DeactiveTeamRepoMap API ==============================
//
func Test_DeactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminDeactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_DeactiveInactiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamRepoMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err == nil {
		t.Errorf("DeactiveInactiveTeamRepoMap Error")
	}
	tearDownTest(t)
}

//
//2.============================== Test ActiveTeamRepoMap API ==============================
//
func Test_ActiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activerepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_SysAdminActiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, sysAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activerepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("PUT", url, nil, sysAdmin); err != nil {
		t.Errorf("ActiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

func Test_RepeatActiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveInactiveTeamRepoMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activerepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeamRepoMap: %v", err.Error())
	}
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err == nil {
		t.Errorf("RepeatActiveTeamRepoMap Error")
	}
	tearDownTest(t)
}

func Test_PushActiveTeamRepoMap(t *testing.T) {
	setUpTest(t)
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Error(err)
	}
	url := setting.ListenMode + "://" + authtest.Domains + "/uam/team/deactiverepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("DELETE", url, nil, orgAdmin); err != nil {
		t.Errorf("DeactiveTeamRepoMap: %v", err.Error())
	}

	url = setting.ListenMode + "://" + authtest.Domains + "/uam/team/activerepository/" + trm.OrgName + "/" + trm.TeamName + "/" + trm.RepoName
	if err := authtest.MethodFunc("PUT", url, nil, orgAdmin); err != nil {
		t.Errorf("ActiveTeamRepoMap: %v", err.Error())
	}
	if err := pushImage(org1, teamAdmin); err != nil {
		t.Errorf("PushActiveTeamRepoMap: %v", err.Error())
	}
	tearDownTest(t)
}

//clear test
func Test_TeamRepostitoryMapClear(t *testing.T) {
	authtest.DeleteUser(t, orgAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, orgMember.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamAdmin.Name, sysAdmin.Name, sysAdmin.Password)
	authtest.DeleteUser(t, teamMember.Name, sysAdmin.Name, sysAdmin.Password)
}
