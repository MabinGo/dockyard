package dao

import (
	"strconv"
	"testing"

	"github.com/huawei-openlab/newdb/orm"
)

func Test_GetPermit(t *testing.T) {
	openDB(t)
	//1. add user
	u := &User{
		Name:     "xiaoming",
		Email:    "xiaoming@huawei.com",
		Password: "xiaoming",
		Status:   0,
		Role:     SYSMEMBER,
	}
	if err := u.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := u.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	//2. add org
	org := &Organization{
		Name:            "huawei",
		MemberPrivilege: WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	//3. add user into org
	orgUserMap := &OrganizationUserMap{
		User: u,
		Role: ORGMEMBER,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Error(err)
	}

	//4. create repo
	repo := &RepositoryEx{
		//Id       int
		Name:     "ubuntu",
		IsPublic: true,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}

	INSERT_TUM_SQL := "insert into  team_user_map(team_id,user_name,role)  values(?,?,?)"
	INSERT_TRM_SQL := "insert into  team_repository_map(team_id,repo_id,permit)  values(?,?,?)"
	o := orm.NewOrm()
	for i := 0; i < 5; i++ {
		//create team
		team := &Team{
			//Id        int           `orm:"auto"`
			Name: "team" + strconv.Itoa(i),
			//Comment   string        `orm:"size(100);null"`
			Org: org,
		}
		if err := team.Save(); err != nil {
			t.Fatal(err)
		}

		if _, err := o.Raw(INSERT_TUM_SQL, team.Id, "xiaoming", TEAMMEMBER).Exec(); err != nil {
			t.Fatal(err)
		}

		if _, err := o.Raw(INSERT_TRM_SQL, team.Id, repo.Id, READ).Exec(); err != nil {
			t.Fatal(err)
		}

		if err := DeactiveTeam(team); err != nil {
			t.Error(err)
		}

		if err := ActiveTeam(team); err != nil {
			t.Error(err)
		}
	}
	if err := DeactiveRepo(repo); err != nil {
		t.Error(err)
	}
	if err := ActiveRepo(repo); err != nil {
		t.Error(err)
	}

	if err := DeactiveOrgUserMap(orgUserMap); err != nil {
		t.Error(err)
	}
	if err := ActiveOrgUserMap(orgUserMap); err != nil {
		t.Error(err)
	}
	if err := DeactiveOrganization(org); err != nil {
		t.Error(err)
	}
	if err := ActiveOrganization(org); err != nil {
		t.Error(err)
	}
	if err := DeactiveUser(u); err != nil {
		t.Error(err)
	}
	if err := ActiveUser(u); err != nil {
		t.Error(err)
	}
	if err := ActiveUser(u); err != nil {
		t.Error(err)
	}

	if permits, err := GetTeamRepoPermit(9, "xiaoming"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(permits)
		for _, p := range permits {
			if p != READ {
				t.Fatal("not pass")
			}
		}
	}
}
