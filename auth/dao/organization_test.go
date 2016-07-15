package dao

import (
	"testing"

	"github.com/containerops/dockyard/utils/db"
)

var openDBFlag bool = false

func openDB(t *testing.T) {

	if openDBFlag {
		return
	}
	if err := db.RegisterDriver("mysql"); err != nil {
		t.Error(err)
	} else {
		db.Drv.RegisterModel(new(Organization), new(User),
			new(OrganizationUserMap), new(RepositoryEx),
			new(Team), new(TeamRepositoryMap), new(TeamUserMap))
		err := db.Drv.InitDB("mysql", "root", "root", "127.0.0.1:3306", "dockyard", 0)
		if err != nil {
			t.Error(err)
		}
	}
	openDBFlag = true
}

func Test_Organization(t *testing.T) {
	openDB(t)
	org := &Organization{
		Name: "huawei",
		//Email           string    `orm:"size(100);null"`
		//Comment         string    `orm:"size(100);null"`
		//URL             string    `orm:"size(100);null"`
		//Location        string    `orm:"size(100);null"`
		MemberPrivilege: WRITE,
		Status:          0,
	}

	if err := org.Save(); err != nil {
		t.Error(err)
	}

	org.Email = "org@gmail.com"
	org.URL = "www.org.com"
	org.Location = "Location"
	org.Comment = "Comment update"
	org.MemberPrivilege = READ
	if err := org.Update("Name", "Email", "Comment", "URL", "Location", "MemberPrivilege", "Updated"); err != nil {
		t.Error(err)
	}

	org1 := &Organization{Name: org.Name}
	if b, err := org1.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("Get Organization error: not found organization")
	} else {
		if org1.Name != org.Name || org1.Email != org.Email || org1.MemberPrivilege != org.MemberPrivilege {
			t.Error("Get Organization error: save is not same with get")
		}
	}
	if err := org.Deactive(); err != nil {
		t.Error(err)
	}
	if _, err := org1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if org1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := org.Active(); err != nil {
		t.Error(err)
	}
	if _, err := org1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if org1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}

	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}
