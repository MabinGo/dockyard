package dao

import (
	"testing"
)

func Test_Team_ExistOrg(t *testing.T) {
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

	team.Comment = "Comment update"
	if err := team.Update("Comment", "Updated"); err != nil {
		t.Fatal(err)
	}

	t1 := &Team{
		//Id        int           `orm:"auto"`
		Name: "team1",
		//Comment   string        `orm:"size(100);null"`
		Org: org,
	}
	if exist, err := t1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("Get team error: not found team")
	} else {
		if t1.Name != team.Name || t1.Comment != team.Comment ||
			t1.Org.Name != team.Org.Name ||
			t1.Org.MemberPrivilege != team.Org.MemberPrivilege {
			t.Error("Get team error: save is not same with get")
		}
	}

	if err := team.Deactive(); err != nil {
		t.Fatal(err)
	}
	if _, err := t1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if t1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := team.Active(); err != nil {
		t.Fatal(err)
	}
	if _, err := t1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if t1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}
}

func Test_Team_InExistOrg(t *testing.T) {
	openDB(t)
	org := &Organization{
		Name: "zte",
		//Email           string    `orm:"size(100);null"`
		//Comment         string    `orm:"size(100);null"`
		//URL             string    `orm:"size(100);null"`
		//Location        string    `orm:"size(100);null"`
		MemberPrivilege: WRITE,
	}

	team := &Team{
		//Id        int           `orm:"auto"`
		Name: "team1",
		//Comment   string        `orm:"size(100);null"`
		Org: org,
	}

	if err := team.Save(); err != nil {
		t.Log(err)
	} else {
		t.Fatal("not pass")
	}

}
