package dao

import (
	"testing"
)

func Test_OrganizationUserMap(t *testing.T) {
	openDB(t)

	u := &User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "111111111",
		//RealName string `orm:"size(100)";null`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   1,
	}
	if err := u.Save(); err != nil {
		t.Error(err)
	}

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

	orgUserMap := &OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: u,
		Role: ORGMEMBER,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Error(err)
	}

	orgUserMap.Role = ORGADMIN
	if err := orgUserMap.Update("Role", "Updated"); err != nil {
		t.Error(err)
	}

	orgUserMap1 := OrganizationUserMap{
		Org:  &Organization{Name: org.Name},
		User: &User{Name: u.Name},
	}

	if b, err := orgUserMap1.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("Get OrganizationUserMap error: not found")
	} else {
		if orgUserMap1.Role != orgUserMap.Role ||
			orgUserMap1.Org.Name != orgUserMap.Org.Name ||
			orgUserMap1.Org.MemberPrivilege != orgUserMap.Org.MemberPrivilege ||
			orgUserMap1.User.Name != orgUserMap.User.Name ||
			orgUserMap1.User.Email != orgUserMap.User.Email ||
			orgUserMap1.User.Status != orgUserMap.User.Status ||
			orgUserMap1.User.Role != orgUserMap.User.Role {
			t.Error("Get OrganizationUserMap error: save is not same with get")
		}
	}
	if err := orgUserMap.Deactive(); err != nil {
		t.Error(err)
	}
	if _, err := orgUserMap1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if orgUserMap1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := orgUserMap.Active(); err != nil {
		t.Error(err)
	}
	if _, err := orgUserMap1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if orgUserMap1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}

	if err := orgUserMap.Delete(); err != nil {
		t.Error(err)
	}

	if err := org.Delete(); err != nil {
		t.Error(err)
	}

	if err := u.Delete(); err != nil {
		t.Error(err)
	}
}
