package dao

import (
	"testing"
)

func Test_TeamRepositoryMap_RepoInOrg(t *testing.T) {
	openDB(t)
	org := &Organization{
		Name: "huawei",
		//Email           string    `orm:"size(100);null"`
		//Comment         string    `orm:"size(100);null"`
		//URL             string    `orm:"size(100);null"`
		//Location        string    `orm:"size(100);null"`
		MemberPrivilege: WRITE,
	}

	if err := org.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	team := &Team{
		//Id        int           `orm:"auto"`
		Name: "team1",
		//Comment   string        `orm:"size(100);null"`
		Org: org,
	}

	if err := team.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := team.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

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
		t.Fatal(err)
	}

	trm := &TeamRepositoryMap{
		//Id      int           `orm:"auto"`
		Team:   team,
		Repo:   repo,
		Permit: WRITE, //team's access permit for repository，
	}
	if err := trm.Save(); err != nil {
		t.Fatal(err)
	}

	trm.Permit = READ
	if err := trm.Update("Permit", "Updated"); err != nil {
		t.Fatal(err)
	}

	trm1 := &TeamRepositoryMap{
		Team: &Team{
			Name: team.Name,
			Org:  &Organization{Name: org.Name},
		},
		Repo: &RepositoryEx{
			Name: repo.Name,
			Org:  &Organization{Name: org.Name},
		},
	}

	if exist, err := trm1.Get(); err != nil {
		t.Fatal(err)
	} else if !exist {
		t.Fatal("Get trm error: not found trm")
	} else {
		if trm1.Permit != trm.Permit {
			t.Fatal("Get trm error: save is not same with get")
		}
	}

	if err := trm.Deactive(); err != nil {
		t.Fatal(err)
	}
	if _, err := trm1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if trm1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := trm.Active(); err != nil {
		t.Fatal(err)
	}
	if _, err := trm1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if trm1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}
}

func Test_TeamRepositoryMap_RepoNotInOrg(t *testing.T) {
	openDB(t)
	org := &Organization{
		Name: "huawei",
		//Email           string    `orm:"size(100);null"`
		//Comment         string    `orm:"size(100);null"`
		//URL             string    `orm:"size(100);null"`
		//Location        string    `orm:"size(100);null"`
		MemberPrivilege: WRITE,
	}

	if err := org.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org1 := &Organization{
		Name: "zte",
		//Email           string    `orm:"size(100);null"`
		//Comment         string    `orm:"size(100);null"`
		//URL             string    `orm:"size(100);null"`
		//Location        string    `orm:"size(100);null"`
		MemberPrivilege: WRITE,
	}
	if err := org1.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := org1.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	team := &Team{
		//Id        int           `orm:"auto"`
		Name: "team1",
		//Comment   string        `orm:"size(100);null"`
		Org: org,
	}

	if err := team.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := team.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	repo := &RepositoryEx{
		//Id       int
		Name:     "ubuntu",
		IsPublic: true,
		//Comment  string
		IsOrgRep: true,
		Org:      org1,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Fatal(err)
	}

	trm := &TeamRepositoryMap{
		//Id      int           `orm:"auto"`
		Team:   team,
		Repo:   repo,
		Permit: WRITE, //team's access permit for repository， 1:write,2:read.
	}
	if err := trm.Save(); err != nil {
		t.Log(err)
	} else {
		t.Fatal("not pass")
	}

}
