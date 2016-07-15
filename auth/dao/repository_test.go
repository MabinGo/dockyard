package dao

import (
	"testing"
)

func Test_Repository(t *testing.T) {
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
		t.Error(err)
	}

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

	repo.IsPublic = false
	repo.Comment = "update test"
	if err := repo.Update("IsPublic", "Comment", "Updated"); err != nil {
		t.Error(err)
	}

	repo1 := &RepositoryEx{
		//Id       int
		Name: repo.Name,
		//IsPublic: repo.IsPublic,
		//Comment  string
		IsOrgRep: repo.IsOrgRep,
		Org:      repo.Org,
		//User     *User
	}
	if b, err := repo1.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("Get repo error: not found repo")
	} else {
		if repo1.Name != repo.Name || repo1.IsPublic != repo.IsPublic ||
			repo1.IsOrgRep != repo.IsOrgRep || repo1.Org.Name != repo.Org.Name ||
			repo1.Org.MemberPrivilege != repo.Org.MemberPrivilege {
			t.Error("Get repo error: save is not same with get")
		}
	}
	if err := repo.Deactive(); err != nil {
		t.Error(err)
	}
	if _, err := repo1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if repo1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := repo.Active(); err != nil {
		t.Error(err)
	}
	if _, err := repo1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if repo1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}

	if err := repo.Delete(); err != nil {
		t.Error(err)
	}

	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}
