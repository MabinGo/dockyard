package dao

import (
	"testing"
)

func Test_TeamUserMap_UserInOrg(t *testing.T) {
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

	u := &User{
		Name:     "rootroot",
		Email:    "liugenping@huawei.com",
		Password: "root",
		//RealName string `orm:"size(100)";null`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   1,
	}
	if err := u.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := u.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: u,
		Role: ORGMEMBER,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Fatal(err)
	}

	tum := &TeamUserMap{
		Team: team,
		User: u,
		Role: TEAMMEMBER,
	}

	if err := tum.Save(); err != nil {
		t.Fatal(err)
	}

	tum.Role = TEAMADMIN
	if err := tum.Update("Role", "Updated"); err != nil {
		t.Fatal(err)
	}

	tum1 := &TeamUserMap{
		Team: &Team{
			Name: team.Name,
			Org:  &Organization{Name: org.Name},
		},
		User: &User{
			Name: u.Name,
		},
	}

	if exist, err := tum1.Get(); err != nil {
		t.Fatal(err)
	} else if !exist {
		t.Fatal("Get tum error: not found tum")
	} else {
		if tum1.Role != tum.Role {
			t.Fatal("Get tum error: save is not same with get")
		}
	}

	if err := tum.Deactive(); err != nil {
		t.Fatal(err)
	}
	if _, err := tum1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if tum1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := tum.Active(); err != nil {
		t.Fatal(err)
	}
	if _, err := tum1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if tum1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}
}

func Test_TeamUserMap_UserNotInOrg(t *testing.T) {
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

	u := &User{
		Name:     "rootroot",
		Email:    "liugenping@huawei.com",
		Password: "root",
		//RealName string `orm:"size(100)";null`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   1,
	}
	if err := u.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := u.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	tum := &TeamUserMap{
		Team: team,
		User: u,
		Role: TEAMMEMBER,
	}

	if err := tum.Save(); err != nil {
		t.Log(err)
	} else {
		t.Fatal("not pass")
	}
}
